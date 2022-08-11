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

import { observer } from 'mobx-react-lite';

import '../../components/dot_spinner';
import { PageLoader } from '../../libs/page_loader';

export interface TestListProps {
  readonly project: string;
  readonly searchQuery: string;
  readonly testLoader: PageLoader<string> | null;
}

export const TestList = observer(({ project, searchQuery, testLoader }: TestListProps) => {
  if (!searchQuery) {
    return <></>;
  }

  if (!testLoader?.loadedFirstPage) {
    return (
      <div>
        Loading
        <milo-dot-spinner></milo-dot-spinner>
      </div>
    );
  }

  return (
    <>
      <ul>
        {testLoader.items.map((testId) => (
          <li key={testId}>
            <a href={`/ui/test/${encodeURIComponent(project)}/${encodeURIComponent(testId)}`} target="_blank">
              {testId}
            </a>
          </li>
        ))}
      </ul>
      {testLoader?.isLoading ? (
        <div>
          Loading
          <milo-dot-spinner></milo-dot-spinner>
        </div>
      ) : (
        <span className="active-text" onClick={() => testLoader?.loadNextPage()}>
          [load more]
        </span>
      )}
    </>
  );
});