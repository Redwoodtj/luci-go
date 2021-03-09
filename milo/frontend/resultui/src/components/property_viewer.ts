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

import { MobxLitElement } from '@adobe/lit-mobx';
import { EditorConfiguration } from 'codemirror';
import { css, customElement, html } from 'lit-element';
import { computed, observable } from 'mobx';

import './code_mirror_editor';
import './connection_observer';
import { ConnectionEvent, ConnectionObserverElement } from './connection_observer';

const LEFT_RIGHT_ARROW = '\u2194';

@customElement('milo-property-viewer')
export class PropertyViewerElement extends MobxLitElement {
  @observable.ref properties!: {[key: string]: unknown};
  @observable.ref foldedLines!: {[key: string]: boolean};
  toggleFold = (_line: string, _fold: boolean) => {};

  @computed private get formattedValue() {
    return JSON.stringify(this.properties, undefined, 2);
  }

  @computed private get formattedValueLines() {
    return this.formattedValue.split('\n');
  }

  private editorOptions: EditorConfiguration = {
    mode: {name: 'javascript', json: true},
    readOnly: true,
    scrollbarStyle: 'null',
    matchBrackets: true,
    lineWrapping: true,
    foldGutter: true,
    lineNumbers: true,
    gutters: ['CodeMirror-linenumbers', 'CodeMirror-foldgutter'],
    foldOptions: {
      widget: (from) => {
        const line = this.formattedValueLines[from.line];
        // Not a root level property, ignore.
        if (!line.startsWith('  "')) {
          return LEFT_RIGHT_ARROW;
        }

        // Use <milo-connection-observer> to observer fold/unfold events.
        // We can't use a regular element with an onclick event handler because
        // code mirror clones the element and all event handlers are dropped.
        // We can't use the 'gutterClick' event because clicking on the widget
        // unfolds the region but doesn't trigger the 'gutterClick' event.
        const connectionObserver = document.createElement('milo-connection-observer') as ConnectionObserverElement<string>;
        connectionObserver.setAttribute('event-type', 'folded-root-lvl-prop');
        connectionObserver.setAttribute('data', JSON.stringify(line));
        connectionObserver.innerHTML = `<span class="CodeMirror-foldmarker">${LEFT_RIGHT_ARROW}<span>`;
        return connectionObserver;
      },
    },
  };

  protected render() {
    return html`
      <milo-code-mirror-editor
        .value=${this.formattedValue}
        .options=${this.editorOptions}
        .onInit=${(editor: CodeMirror.Editor) => {
          this.formattedValueLines.forEach((line, lineIndex) => {
            if (this.foldedLines[line]) {
              editor.foldCode(lineIndex);
            }
          });
        }}
        @folded-root-lvl-prop=${(e: ConnectionEvent<string>) => {
          if (!this.foldedLines[e.detail.data]) {
            this.toggleFold(e.detail.data, true);
          }

          e.detail.addDisconnectedCB((data) => {
            // If the widget is disconnected because the property
            // viewer is disconnected, ignore.
            if (this.isConnected && this.foldedLines[data]) {
              this.toggleFold(data, false);
            }
          });
        }}
      ></milo-code-mirror-editor>
    `;
  }

  static styles = css`
    :host {
      display: block;
    }
  `;
}