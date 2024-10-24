// Copyright 2022 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package quota provides an implementation for server quotas which are backed
// by Redis.
//
// # Rationale
//
// Quotas are a way to restrict shared resource consumption in order to provide
// fairness and prevent abuse. The quota library implements a way to configure
// and track resource limits for users for application-specific resources.
//
// We intend that this library be a 'good enough' implementation that it can
// serve the needs of many (if not all) LUCI services and provide additional
// common benefits (logging, metrics, administration ACLs/API/UI) so that each
// individual service doesn't need to re-invent these mechanisms.
//
// The current implementation is based on Redis and is fully synchronous.
// There's a possibility in the future that we could extend the implementation
// to other datastores or to allow the application to make a tradeoff between
// accuracy and latency.
//
// # Data Model
//
// There are 2 different types of entities managed by the quota libary: Policies
// (grouped into a PolicyConfig) and Accounts. The library provides a variety of
// Operations which all work in terms of these entities.
//
// # Data Model - Entity identities
//
// All entities have an identity which is composed of the following 'atoms'.
// Some of these atoms need structure which is meaningful to the application.
// The quota library has a convention for such atoms called "ASIs" (Application
// specific identifiers). See that section for what/why. Note that all of these
// identifiers end up as Redis keys (or hash keys) one way or the other, so all
// the usual caveats around absurd key lengths apply here. However, Redis allows
// keys up to 512MB, so have fun...
//
// Common identifier atoms:
//
//  - app_id - The app_id allows multiple logical applications to share the same
//    Redis instance. This should reflect the service that the account or policy
//    belongs to. For example this would allow a single deployment to have quota
//    accounts/policies for an application "cv" and "rdb" in the same binary.
//  - realm - For administration purposes, Accounts and PolicyConfigs belong to
//    a realm (though likely not the same one). Typically, PolicyConfigs will
//    belong to a project's @project realm. Accounts will belong to realms which
//    make sense in the context of the application. `realm` here is a global
//    realm (i.e. `project:something`).
//  - resource_type (ASI) - A given Policy or Account can only deal in a single
//    resource_type. This value only needs to make sense to the application.
//  - namespace (ASI) - Namespace allows the Application to segment a given
//    realm into multiple sub-domains. For example, Buildbucket could use the
//    namespace to indicate that a given Account is being used for a single
//    builder within a bucket. This only needs to make sense to the application.
//  - name (ASI) - Name is the name of the entity. This only needs to make sense
//    to the application.
//
// # Data Model - PolicyConfig
//
// ID: app_id ~ realm ~ version
//
// A PolicyConfig is an immutable group (Redis Hash) of Policies.
// Typically this will be in a @project realm of some LUCI project, as current
// users will likely derive a PolicyConfig from some other LUCI project
// configuration.
//
// The realm indicates which realm this PolicyConfig is administered under, but
// it doesn't need to (and likely will not) match the realm for Accounts using
// the Policies within it.
//
// In the PolicyConfig ID, the `version` field is a content hash (starting with
// `$`), or manually supplied ("#" followed by an ASI). Once written,
// PolicyConfigs cannot be modified (but they can be purged). It's recommended
// to use the content hash versioning scheme (this will also do implicit
// deduplication when configs change without policy changes). However, some
// applications may find it more convenient to tie the PolicyConfig version to
// an external version identifier (like a git commit id of the overall configs),
// so manually versioning the PolicyConfigs is an option.
//
// Purging PolicyConfigs results in the deletion of a PolicyConfig and should
// only be used for PolicyConfigs that the application knows are no longer in
// use. However, in the event that a PolicyConfig is purged while Accounts still
// list it:
//   * Operations on those Accounts without supplying a new Policy reference
//     will continue to use the snapshot of the policy stored in the Account.
//     We could potentially make this produce a warning or error, however.
//   * Operations on Accounts that supply a new Policy reference must have that
//     Policy exist, as usual, and it will replace the referenced/snapshotted
//     policy in the Account.
//
// # Data Model - Policy
//
// Key (within a PolicyConfig): namespace ~ name ~ resource_type
//
// A Policy is an immutable member of a PolicyConfig, and stores a numeric
// Default, Limit, Refill, and a Lifetime.
//   - Default - The value to set a previously non-existant Account to when
//     first accessing it.
//   - Limit - The maximum value an Account can have.
//   - Lifetime - The number of seconds to wait before garbage collecting an
//     Account after its last update. This is implemented with a Redis TTL which
//     is refreshed on the Account each time it's written.
//
// Refill is a numeric triple (see the "Refill Behavior" section for details of
// how refill works):
//   - Units - The number of units to add.
//   - Interval - The number of seconds in between fill events. Intervals are
//     synchronized to UTC midnight + Offset. See the "Refill Behavior" section
//     for a discussion on how Refill is implemented. Note that there is no cron
//     or "stampede" from synchronizing refill events in this way.
//   - Offset - The number of seconds to offset UTC midnight to the 0th daily
//     interval.
//
// # Data Model - Account
//
// ID: app_id ~ realm ~ namespace ~ name ~ resource_type
//
// Accounts hold the balance of a specific owning identity for a specific
// resource. They contain:
//   - Balance - Current number of units held.
//   - LastUpdate - Time when this Account was last updated.
//   - LastRefill - Time when this Account was last refilled (always <=
//     LastUpdate).
//   - LastPolicyChange - Time when the currently applied Policy was first
//     set.
//   - Options - Bit field indicating various options. Currently the only option
//     is `MANUAL_WRITE_OK=1` which enables `quota.accounts.write` for this
//     account.
//   - PolicyConfig - Redis key for the versioned PolicyConfig last used for this
//     Account.
//   - PolicyKey - Hash key (namespace ~ name ~ resource_type) in the PolicyConfig
//     for the Policy last used for this Account.
//   - PolicyRaw - Raw encoded snapshot of the last-used policy for this Account.
//     This is necessary to allow the quota library to interact with an Account
//     under it's last-applied policy without needing to re-read the original
//     policy (which is technically difficult to do in Redis scripts because
//     they need to have all Redis keys supplied to them in advance of their
//     execution).
//
// # Operations
//
// Operations combine a Policy with an Account in one of a few pre-defined ways:
//   - Delta - make a positive or negative adjustment to an Account balance
//     constrained by a Policy.
//   - Reset - set the Account balance to one of the constants in the Policy
//     (i.e. Default, Zero or Limit).
//   - Set - set the Account balance to an arbitrary value (optional Policy
//     will cap the value)
//
// All operations allow supplying or omitting a Policy. If omitted, will use the
// policy details stored in the Account. If supplied, this Policy becomes
// the stored Policy for the account.
//
// When an Operation occurs, the system first does a Refill (see the "Refill
// Behavior" section), followed by the application of the Operation.
//
// It is permitted to apply multiple Operations (possibly to differing accounts
// under different policies) in the same RPC.
//
// All Operations will also return the after-adjustment-balance.
//
// There will also be a Get operation which ONLY reads the data, returning the
// full Account data and also the projected value (e.g. after refills). This
// operation does NOT change the LastRefill timestamp.
//
// # Application-specific identifiers (ASIs)
//
// The quota library has several application-specific identifiers (ASIs). These
// ASIs end up ~verbatim in Redis as row keys. This means that your storage
// costs and lookup performance will be proportional to their length.
//
// The quota libary reserves the character "~" for partitioning ASIs when
// synthesizing a full Redis key.
//
// Additionally, two characters will be treated specially as a convention:
//   - "|" is available to separate sections within an ASI.
//   - "{", if the first character in an ASI section, indicates that the
//     remainder of that section is encoded with ascii85 (an encoding which
//     conveniently excludes "~", "|", and "{"). Functions in this library
//     which attempt to do this interpretation will return the raw string
//     instead of failing (e.g. if you had `{z` in a section, it would be
//     returned as `{z` rather than as an error).
//
// The quota library provides functions to encode/decode a series of arbitrary
// section strings to/from a single ASI string.
//
// The quota library may use "|" as a way to group related keys together when
// displaying a large collection of quota Account or Policy data. Think of it
// similarly to how GCS treats "/". It's a visual delimiter, but the underlying
// service doesn't really care if you use it or not. Similarly, sections
// starting with '{' will attempt to decode in certain contexts (like the UI),
// but if decoding fails it will return the original string. If your application
// dosen't care about this functionality at all, it's free to use any string it
// likes as an ASI, as long as it doesn't contain `~`.
//
// # Refill Behavior
//
// Refills in the quota library are intended to mimic the behavior of a cron job
// which runs every second, scanning all Accounts, seeing if their Interval is
// past and refilling them.
//
// However, such an implementation would be terribly slow. Instead, the quota
// library remembers the policy details for each account and then when
// interacting with the Account as part of an Operation, this will refill based
// on the real elapsed time under the previous Policy.
//
// Refills are synchronized to UTC plus an offset. This means if you specify 17
// units with an interval of "21600" (i.e. 6 hours), and an offset of 0, then
// each 6 hours after UTC midnight, 17 units would be added to the account. If
// the account was created at, say, 0740 UTC, then the next refill event would
// occur at 1200 UTC.
//
// Offset allows you to 'rotate' this cycle so that a given policy's "midnight"
// occurs at a different time of day. (NOTE: Theoretically this offset could be
// per-Account rather than per-Policy. If this becomes a necessary usecase, it
// wouldn't be hard to add, but for now we're keeping it simple).
//
// Please also refer to "Implementation notes - Refill Interval" and
// "Implementation notes - Refill Synchronization" for a discussion on why we
// picked this Refill system vs. a simpler units/second alternative and why we
// tie refills to the wall clock time.
//
// # Behavior when switching Policies
//
// Over time, it is likely that a single Account will go through multiple
// different Policies which apply to it, or where those Policies change
// parameters over time.
//
// Account names should always be stable, comprising a who/what/where of
// a resource. When policies shift for an Account, the quota library will
// maintain the previous balance of the Account, except that no Refill will take
// place if the Account is over its limit. Additionally, no matter how far out
// of spec an Account is, it will always be permitted to make an over-limit
// account smaller, or an under-zero account larger.
//
// So, say an account had a policy which had a limit of 20, with a balance of
// 18, and switched to a policy with a balance of 15. It would maintain its
// balance of 18 until debited, but any positive refill policy would have no
// effect.
//
// # Access control and Administration
//
// The quota library implements an administration service API. This is an
// auxilliary API to read/write the values manipulated by the quota library, to
// be used for debugging or manual intervention (rather than directly poking the
// underlying Redis data).
//
// Access via this service is granted via realm permissions:
//   - quota.accounts.read - Allows reading single accounts within a realm.
//     Binding context: {app_id, resource_type, namespace}
//   - quota.accounts.list - Allows listing accounts
//     Binding context: {app_id, resource_type, namespace}
//   - quota.accounts.write - Allows modifying accounts. Note that this only
//     applies to accounts which have the option MANUAL_WRITE_OK. For example,
//     some Applications use Accounts to represent a number of active
//     resources; Allowing users to directly change this value via the Admin API
//     doesn't actually make sense.
//     Binding context: {app_id, resource_type, namespace}
//   - quota.policies.read - Allows reading policy contents.
//     Binding context: {app_id}
//   - quota.policies.write - Allows writing new content-addressed policy
//     configs. Binding context: {app_id}
//   - quota.policies.overrideVersion - If granted in conjunction with
//     `quota.policies.write`, allows writing new manually-versioned policy
//     configs. Binding context: {app_id}. Note that manually-versioned policy
//     configs are not verifiable by the quota library and could allow users
//     with this permission to 'poison' a quota policy version.
//   - quota.policies.purge - Allows perging PolicyConfigs.
//     Binding context: {app_id}.
//
// Permission checks require one of:
//   - hasPermission(perm, operation_realm) OR
//   - hasPermission(perm, "@internal:<service-app-id>")
//
// That is, internal permissions can be granted to service deployment
// Admins.
//
// NOTE: These access controls ONLY apply to requests via the Administration
// service API. Interaction with the quotas via the Go API do not do any access
// checking, because it is assumed that the application has already done
// appropriate access checks before computing the Accounts/Policies to interact
// with.
//
// # Implementation notes - Refill Interval
//
// Initially the Quota library implemented a "units/second" refill system. This
// made the implementation nice due to its simplicity, but had two noticeable
// drawbacks:
//
//   1. Low quantity quotas (e.g. builds per day) were difficult to express
//      naturally (for example, the application would have to have accounts in
//      fractional builds, like 100,000 == one build).
//   2. Even if the application expressed account values in this way, this leads
//      to an effectively "analog" replenismhent system which would lead to
//      mistakes when setting quotas.
//
// Consider the case where you want to restrict users to "10 builds per day".
// You first make the accounts hold thousandths of a build, and then set
// a policy with (limit=1000000, refill_each_sec=11). Ignoring the fact that the
// refill should actually be something like 11.574, we've basically achieved
// what we want, right? A user can only run 10 builds (a bit less) per day.
//
// Not quite. Consider that the user can wait until their quota is full (10
// builds) and then they:
//   * Run 10 builds in hour 0
//   * Run one build every ~2 hours for the next 24 hours.
//
// Oops... our 10/day quota actually allows the user to burst up to 19/day.
// Mondays are gonna be spicy.
//
// # Implementation notes - Refill Synchronization
//
// Quota refills are tricky; originally we started the clock at account creation
// time, but realized this would lead to two issues:
//
//   1. Every quota account would refresh at seemingly-random times, which makes
//      debugging more difficult. This would not be beneficial for 'load
//      distribution' in a system (it should explicitly use short term quotas or
//      some othe rate limiting techniques instead).
//   2. This would lead to very difficult to reason-about behaviors when
//      policies change for a given account.
//
// In the case of policy changes, the only sensible thing to do while
// maintaining the interval based refill events would be to reset the refill
// timer when changing policies on an account. However, for Refill policies with
// long intervals, this could lead to artifacts where users are inexplicably
// starved for quota. Consider a situation where a user is allowed 10 builds per
// day. They exhaust their quota at hour 23 of the day and complain to a trooper
// who then moves them to a higher-tier policy group with 20 builds per day.
//
// However, when hour 24 rolls around, the user's account not only doesn't get
// 20 builds added to it, it doesn't even get the original 10. Instead the user
// has to wait an ADDITIONAL 24h before their quota replenishes.
//
// Synchronizing refill events significantly improves the predictability of the
// system here.
//
// # Implementation notes - Deduplication
//
// The quota library has a simple deduplication scheme which is indended to
// prevent accidentally applying Operations multiple times (for example,
// applying a Delta(-10) operation twice when you only wanted to apply it once
// could be pretty bad).
//
// When any actor interacts with the Quota library (either via the Go interface
// or the Administration API), they provide a request ID. The quota library then
// calculates if ALL of the Operations in the request can proceed with the
// current Account state, and, if so, applies ALL of the Operations atomically*,
// followed by recording the RequestID into Redis with a TTL (defaulting to
// 2 hours), a hash of the requested operations, plus the returned value for the
// Account balances after applying all of the Operations. If a subsequent
// request comes in with the same RequestID, the hash of the Operations is
// checked, and if it matches the stored value, the original result will be
// returned without error.
//
// (* I put the scary asterisk on atomically, because _as far as I can tell_,
// EVAL scripts in Redis are either fully applied, or not applied at all.
// However the statements in the docs aren't as strong as I'd like to this
// effect. The docs do state that EVAL (or FUNCTIONs) is our best bet.)
//
// Supplying a different set of Operations with the same RequestID is an error,
// and the request will be rejected.
//
// Where this departs from "normal" deduplication is that _negative_ (error)
// results are NOT recorded; That is, if you attempt to debit an account "A"
// by 1 unit, but the balance is currently 0, this will return an "underflow"
// error, but the RequestID will not be consumed (so retrying this exact same
// request later may succeed, if the balance of "A" has risen above 1.
//
// We speculate that this mode is more intuitive, since many of the places we
// expect applications to interact with the quota library are attempting to make
// rapid, otherwise stateless, decisions about what to do next, where generating
// the RequestID deterministically in the context of that decision is
// convenient. If we stored the rejection via the RequestID, it would require
// these stateless invocations to likely store the fact that a RequestID was
// consumed, or to pick randomized RequestIDs (which then gets you in trouble
// when multiple processes are attempting to make the same decision and would
// only fail out on a transaction after communicating intent to the quota
// service).
//
// # Implementation notes - Redis encoding
//
// This library makes use of `msgpack` to encode both Accounts and Policies in
// Redis. Unfortunately, because we need to implement quota manipulation in
// `lua`, protobuf wasn't an option for these.
//
// Data in Redis is stored using a version-prefixed "array" encoding of the
// structure. For example, an Account like:
//
//   {
//     "balance": 1938,
//     "last_update": 1662142122,
//     "last_refill": 1662076800,
//     "last_policy_change": 1660439785,
//     "options": 1,
//     "policy_key": "cv~chromium:@project~$mo+]^aHaN<Jl6//WO=sVWaccSe^oYp[q.sdL,JjE",
//     "policy_name": "policyname",
//     "policy_limit": 10000,
//     "policy_refill": {
//       "units": 100,
//       "interval": 600,
//       "offset": 0
//     }
//   }
//
// Will currently be encoded like (the prepended 0s are version identifiers for
// the remaining encoded object. If we need to change the schema we can use this
// to switch between decoders):
//
//   msgpack(0) +
//   msgpack([
//     1938,
//     1662142122,
//     1662076800,
//     1660439785,
//     1,
//     "cv~chromium:@project~$mo+]^aHaN<Jl6//WO=sVWaccSe^oYp[q.sdL,JjE",
//     "policyname",
//     10000,
//     [
//       100,
//       600,
//       0
//     ]
//   ])
//
// Which is 105 bytes that looks like:
//
//   00 99 cd 07 92 ce 63 12 46 aa ce 63 11 47 80 ce 62 f8 4c e9 01 d9 3e 63 76 7e
//   63 68 72 6f 6d 69 75 6d 3a 40 70 72 6f 6a 65 63 74 7e 24 6d 6f 2b 5d 5e 61 48
//   61 4e 3c 4a 6c 36 2f 2f 57 4f 3d 73 56 57 61 63 63 53 65 5e 6f 59 70 5b 71 2e
//   73 64 4c 2c 4a 6a 45 aa 70 6f 6c 69 63 79 6e 61 6d 65 cd 27 10 93 64 cd 02 58
//   00
//
// (Note that a JSON object encoding (the other encoder available in Redis) of
// the same would be roughly 247 bytes or more. Doing a JSON array encoding is
// closer (190 or so), but still less efficient in terms of storage and
// encoding/decoding.)
//
// Assuming account names of ~64 bytes, this means that a single account can be
// estimated to take 256 bytes or less, which would allow a single 3GB Redis
// instance to pessimistically manage a bit over 10 million active quota
// accounts before starting evictions (leaving > 128MB for additional overhead).
//
// It would be possible to further shrink the Account by creating a shortened
// form of the policy_ref, which takes 76 of the 117 bytes, roughly doubling the
// number of manageable accounts. This could be done by ingesting new policy
// config versions and assigning them a numeric ID, as well as assigning
// a numeric ID foR the policy name. Then representing the policy_ref would take
// no more than 10 bytes (assuming 32bit representations were required for both,
// which is a stretch).
package quota
