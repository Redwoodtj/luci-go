// Copyright 2021 The LUCI Authors.
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

/**
 * @fileoverview
 * This file contains utility types that help us mock the interface of the
 * improved test history API.
 *
 * TODO(crbug/1266759): once we have the improved test history API, switch to it.
 */

import escapeStringRegexp from 'escape-string-regexp';
import stableStringify from 'fast-json-stable-stringify';

import { cached, CacheOption } from '../libs/cached_fn';
import * as iterUtils from '../libs/iter_utils';
import { parseProtoDuration } from '../libs/time_utils';
import {
  GetTestResultHistoryRequest,
  GetTestResultHistoryResponseEntry,
  ResultDb,
  TestExoneration,
  TestResult,
  TestStatus,
  TestVariantStatus,
  TimeRange,
  Variant,
} from './resultdb';

export interface QueryTestHistoryRequest {
  readonly realm: string;
  readonly testId: string;
  readonly pageSize?: number;
  readonly pageToken?: string;
  readonly timeRange: TimeRange;
}

export interface TestVariantSummary {
  readonly testId: string;
}

export interface TestVariantHistoryEntry {
  readonly invocationTimestamp: string;
  readonly variant?: Variant;
  readonly variantHash: string;
  readonly status: TestVariantStatus;
  readonly averageDuration?: string;
}

export interface QueryTestHistoryResponse {
  readonly entries: readonly TestVariantHistoryEntry[];
  readonly nextPageToken?: string;
}

/**
 * A service that mocks the interface of the improved test history API.
 */
export class TestHistoryService {
  constructor(private readonly resultdb: ResultDb) {}

  /**
   * Create an generator that streams test result history response entries from
   * the current test history API.
   */
  private async *genTestResultHistoryEntry(
    realm: string,
    testId: string,
    timeRange: TimeRange
  ): AsyncGenerator<GetTestResultHistoryResponseEntry, void, undefined> {
    // '' -> first page, null -> no more pages.
    let nextPageToken: string | null = '';
    const nextTimeRange = { ...timeRange };

    while (nextPageToken !== null) {
      const getTRHReq: GetTestResultHistoryRequest = {
        realm: realm,
        testIdRegexp: escapeStringRegexp(testId),
        timeRange: nextTimeRange,
        pageToken: nextPageToken,
      };
      const res = await this.resultdb.getTestResultHistory(getTRHReq, { skipUpdate: true });

      // This is the last page, yield every entries.
      if (!res.nextPageToken) {
        nextPageToken = null;
        yield* res.entries;
        continue;
      }

      const lastEntry = res.entries[res.entries.length - 1];

      // All the entries have the same timestamp. We can't emulate pagination by
      // advancing the time range.
      if (res.entries[0].invocationTimestamp === lastEntry.invocationTimestamp) {
        nextPageToken = res.nextPageToken;
        yield* res.entries;
        continue;
      }

      // The current test history API's pagination implementation is slow and
      // may lead to timeout. Emulate pagination by advancing the time range.
      nextTimeRange.latest = lastEntry.invocationTimestamp;
      nextPageToken = '';
      // Discard results with the same timestamp as nextTimeRange.latest.
      // They will be returned again in the next RPC call.
      yield* res.entries.filter((e) => e.invocationTimestamp !== lastEntry.invocationTimestamp);
    }
  }

  /**
   * Create a generator that create test variant history entries and yield them.
   */
  private async *genTestVariantHistoryEntry(realm: string, testId: string, timeRange: TimeRange) {
    const trhEntryIter = this.genTestResultHistoryEntry(realm, testId, timeRange);

    let results: TestResult[] = [];
    let tvhEntry: DeepMutable<TestVariantHistoryEntry> | null = null;

    for await (const trh of trhEntryIter) {
      const sameTVHEntry =
        tvhEntry !== null &&
        trh.invocationTimestamp === tvhEntry.invocationTimestamp &&
        trh.result.variantHash === tvhEntry.variantHash;

      // Result entry belongs to the same test variant history entry, add it to
      // the last entry.
      if (sameTVHEntry) {
        results!.push(trh.result);
        continue;
      }

      // Result entry belongs to a different test variant, finalize the last
      // entry and yield it.
      if (tvhEntry !== null) {
        yield finalizedTVEntry(tvhEntry, results);
      }

      // Then create a new entry.
      tvhEntry = {
        invocationTimestamp: trh.invocationTimestamp,
        variant: trh.result.variant,
        variantHash: trh.result.variantHash || '',
        status: TestVariantStatus.TEST_VARIANT_STATUS_UNSPECIFIED,
      };
      results = [trh.result];
    }

    // Finalize the last entry and yield it.
    if (tvhEntry !== null) {
      yield finalizedTVEntry(tvhEntry, results);
    }
  }

  /**
   * Saved iterators for pagination purpose.
   */
  private readonly pageIterCache: {
    [key: string]: () => AsyncIterableIterator<TestVariantHistoryEntry>;
  } = {};

  private queryTestHistoryImpl = async (req: QueryTestHistoryRequest): Promise<QueryTestHistoryResponse> => {
    // Recover the iterator when page token is used.
    const iter = req.pageToken
      ? this.pageIterCache[req.pageToken]()
      : this.genTestVariantHistoryEntry(req.realm, req.testId, req.timeRange);

    const entries: TestVariantHistoryEntry[] = [];
    let remaining = req.pageSize || 100;
    let reachedEnd = false;

    // Get up to req.pageSize number of entries from the iterator.
    //
    // We may need to save the iterator later.
    // Don't use a for-await-of loop because it will terminate the generator.
    // https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/for...of#do_not_reuse_generators
    while (remaining > 0) {
      const next = await iter.next();
      if (next.done) {
        reachedEnd = true;
        break;
      }
      entries.push(next.value);
      remaining--;
    }

    const res: Mutable<QueryTestHistoryResponse> = {
      entries: entries,
    };

    // If we haven't reached the end of the list, generate a random page token
    // and save the iterator.
    if (!reachedEnd) {
      res.nextPageToken = Math.random().toString();
      // Tee the iter so it can be reused when the caller requested the same
      // page.
      // TODO(crbug/1266759): this implementation will cause memory leaks but we
      // accept this since the implementation is only temporary.
      this.pageIterCache[res.nextPageToken] = iterUtils.teeAsync(iter);
    }

    return res;
  };

  private cachedQueryTestHistoryImpl = cached(this.queryTestHistoryImpl, { key: (req) => stableStringify(req) });

  async queryTestHistory(req: QueryTestHistoryRequest, cacheOpt: CacheOption = {}) {
    return this.cachedQueryTestHistoryImpl(cacheOpt, req);
  }
}

/**
 * Given a list of test results, return the status that their parent test
 * variant entry should have.
 */
function computedTestVariantStatus(
  results: readonly TestResult[],
  exonerations: readonly TestExoneration[]
): TestVariantStatus {
  if (exonerations.length !== 0) {
    return TestVariantStatus.EXONERATED;
  }

  let hasExpected = false;
  let hasUnexpected = false;
  let allUnexpectedWereSkipped = true;

  for (const result of results) {
    if (result.expected) {
      hasExpected = true;
    } else {
      hasUnexpected = true;
      if (result.status !== TestStatus.Skip) {
        allUnexpectedWereSkipped = false;
      }
    }

    if (hasExpected && hasUnexpected) {
      break;
    }
  }

  if (hasExpected) {
    if (hasUnexpected) {
      return TestVariantStatus.FLAKY;
    }
    return TestVariantStatus.EXPECTED;
  }
  if (allUnexpectedWereSkipped) {
    return TestVariantStatus.UNEXPECTEDLY_SKIPPED;
  }
  return TestVariantStatus.UNEXPECTED;
}

/**
 * Compute the status and the average duration for the test variant history
 * entry.
 */
function finalizedTVEntry(tvhEntry: DeepMutable<TestVariantHistoryEntry>, results: TestResult[]) {
  tvhEntry.status = computedTestVariantStatus(results, []);
  const durations = results.filter((r) => r.duration).map((r) => parseProtoDuration(r.duration!));
  if (durations.length > 0) {
    tvhEntry!.averageDuration = `${durations.reduce((a, b) => a + b, 0) / durations.length}s`;
  }
  return tvhEntry;
}