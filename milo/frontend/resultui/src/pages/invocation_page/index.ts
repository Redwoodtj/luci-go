// Copyright 2020 The LUCI Authors.
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

import '@material/mwc-icon';
import { BeforeEnterObserver, PreventAndRedirectCommands, RouterLocation } from '@vaadin/router';
import { css, customElement, html } from 'lit-element';
import { computed, makeObservable, observable, reaction } from 'mobx';

import '../../components/status_bar';
import '../../components/tab_bar';
import '../test_results_tab/count_indicator';
import { MiloBaseElement } from '../../components/milo_base';
import { TabDef } from '../../components/tab_bar';
import { InvocationState, provideInvocationState } from '../../context/invocation_state';
import { INVOCATION_STATE_DISPLAY_MAP } from '../../libs/constants';
import { consumer, provider } from '../../libs/context';
import { reportRenderError } from '../../libs/error_handler';
import { NOT_FOUND_URL, router } from '../../routes';
import { consumeStore, StoreInstance } from '../../store';
import commonStyle from '../../styles/common_style.css';

/**
 * Main test results page.
 * Reads invocation_id from URL params.
 * If not logged in, redirects to '/login?redirect=${current_url}'.
 * If invocation_id not provided, redirects to '/not-found'.
 * Otherwise, shows results for the invocation.
 */
@customElement('milo-invocation-page')
@provider
@consumer
export class InvocationPageElement extends MiloBaseElement implements BeforeEnterObserver {
  @observable.ref
  @consumeStore()
  store!: StoreInstance;

  @observable.ref
  @provideInvocationState({ global: true })
  invocationState!: InvocationState;

  constructor() {
    super();
    makeObservable(this);
  }

  private invocationId = '';
  onBeforeEnter(location: RouterLocation, cmd: PreventAndRedirectCommands) {
    const invocationId = location.params['invocation_id'];
    if (typeof invocationId !== 'string') {
      return cmd.redirect(NOT_FOUND_URL);
    }
    this.invocationId = invocationId;
    return;
  }

  connectedCallback() {
    super.connectedCallback();

    this.addDisposer(
      reaction(
        () => [this.store],
        ([store]) => {
          this.invocationState?.dispose();
          this.invocationState = new InvocationState(store);
          this.invocationState.invocationId = this.invocationId;

          // Emulate @property() update.
          this.updated(new Map([['invocationState', this.invocationState]]));
        },
        { fireImmediately: true }
      )
    );
    this.addDisposer(() => this.invocationState.dispose());

    document.title = `inv: ${this.invocationId}`;
  }

  private renderInvocationState() {
    const invocation = this.invocationState.invocation;
    if (!invocation) {
      return null;
    }
    if (invocation.finalizeTime) {
      return html`
        <i>${INVOCATION_STATE_DISPLAY_MAP[invocation.state]}</i>
        at ${new Date(invocation.finalizeTime).toLocaleString()}
      `;
    }

    return html`
      <i>${INVOCATION_STATE_DISPLAY_MAP[invocation.state]}</i>
      since ${new Date(invocation.createTime).toLocaleString()}
    `;
  }

  @computed get tabDefs(): TabDef[] {
    return [
      {
        id: 'test-results',
        label: 'Test Results',
        href: router.urlForName('invocation-test-results', { invocation_id: this.invocationState.invocationId! }),
        slotName: 'test-count-indicator',
      },
      {
        id: 'invocation-details',
        label: 'Invocation Details',
        href: router.urlForName('invocation-details', { invocation_id: this.invocationState.invocationId! }),
      },
    ];
  }

  protected render = reportRenderError(this, () => {
    if (this.invocationState.invocationId === '') {
      return html``;
    }

    return html`
      <div id="test-invocation-summary">
        <div id="test-invocation-id">
          <span id="test-invocation-id-label">Invocation ID </span>
          <span>${this.invocationState.invocationId}</span>
        </div>
        <div id="test-invocation-state">${this.renderInvocationState()}</div>
      </div>
      <milo-status-bar
        .components=${[{ color: 'var(--active-color)', weight: 1 }]}
        .loading=${this.invocationState.invocation === null}
      ></milo-status-bar>
      <milo-tab-bar .tabs=${this.tabDefs} .selectedTabId=${this.store.selectedTabId}>
        <milo-trt-count-indicator slot="test-count-indicator"></milo-trt-count-indicator>
      </milo-tab-bar>
      <slot></slot>
    `;
  });

  static styles = [
    commonStyle,
    css`
      #test-invocation-summary {
        background-color: var(--block-background-color);
        padding: 6px 16px;
        font-family: 'Google Sans', 'Helvetica Neue', sans-serif;
        font-size: 14px;
        display: flex;
      }

      milo-tab-bar {
        margin: 0 10px;
        padding-top: 10px;
      }

      milo-trt-count-indicator {
        margin-right: -13px;
      }

      #test-invocation-id {
        flex: 0 auto;
      }

      #test-invocation-id-label {
        color: var(--light-text-color);
      }

      #test-invocation-state {
        margin-left: auto;
        flex: 0 auto;
      }
    `,
  ];
}
