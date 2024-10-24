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

import fetchMock from 'fetch-mock-jest';

import {
  BatchGetClustersRequest,
  BatchGetClustersResponse,
  Cluster,
  ClusterExoneratedTestVariant,
  ClusterSummary,
  DistinctClusterFailure,
  QueryClusterFailuresRequest,
  QueryClusterFailuresResponse,
  QueryClusterSummariesRequest,
  QueryClusterSummariesResponse,
  QueryClusterExoneratedTestVariantsRequest,
  QueryClusterExoneratedTestVariantsResponse,
} from '@/services/cluster';

export const getMockCluster = (id: string,
    project = 'testproject',
    algorithm = 'reason-v2',
    title = ''): Cluster => {
  return {
    'name': `projects/${project}/clusters/${algorithm}/${id}`,
    'hasExample': true,
    'title': title,
    'userClsFailedPresubmit': {
      'oneDay': { 'nominal': '98' },
      'threeDay': { 'nominal': '158' },
      'sevenDay': { 'nominal': '167' },
    },
    'criticalFailuresExonerated': {
      'oneDay': { 'nominal': '5625' },
      'threeDay': { 'nominal': '14052' },
      'sevenDay': { 'nominal': '13800' },
    },
    'failures': {
      'oneDay': { 'nominal': '7625' },
      'threeDay': { 'nominal': '16052' },
      'sevenDay': { 'nominal': '15800' },
    },
    'equivalentFailureAssociationRule': '',
  };
};

export const getMockRuleClusterSummary = (id: string): ClusterSummary => {
  return {
    'clusterId': {
      'algorithm': 'rules-v2',
      'id': id,
    },
    'title': 'reason LIKE "blah%"',
    'bug': {
      'system': 'buganizer',
      'id': '123456789',
      'linkText': 'b/123456789',
      'url': 'https://buganizer/123456789',
    },
    'presubmitRejects': '27',
    'criticalFailuresExonerated': '918',
    'failures': '1871',
  };
};

export const getMockSuggestedClusterSummary = (id: string, algorithm = 'reason-v3'): ClusterSummary => {
  return {
    'clusterId': {
      'algorithm': algorithm,
      'id': id,
    },
    'bug': undefined,
    'title': 'reason LIKE "blah%"',
    'presubmitRejects': '29',
    'criticalFailuresExonerated': '919',
    'failures': '1872',
  };
};

export const getMockClusterExoneratedTestVariant = (id: string, exoneratedFailures: number): ClusterExoneratedTestVariant => {
  return {
    'testId': id,
    'criticalFailuresExonerated': exoneratedFailures,
    'lastExoneration': '2052-01-02T03:04:05.678901234Z',
  };
};

export const mockQueryClusterSummaries = (request: QueryClusterSummariesRequest, response: QueryClusterSummariesResponse) => {
  fetchMock.post({
    url: 'http://localhost/prpc/luci.analysis.v1.Clusters/QueryClusterSummaries',
    body: request,
  }, {
    headers: {
      'X-Prpc-Grpc-Code': '0',
    },
    body: ')]}\'' + JSON.stringify(response),
  }, { overwriteRoutes: true });
};

export const mockBatchGetCluster = (
    project: string,
    algorithm: string,
    id: string,
    responseCluster: Cluster) => {
  const request: BatchGetClustersRequest = {
    parent: `projects/${encodeURIComponent(project)}`,
    names: [
      `projects/${encodeURIComponent(project)}/clusters/${encodeURIComponent(algorithm)}/${encodeURIComponent(id)}`,
    ],
  };

  const response: BatchGetClustersResponse = {
    clusters: [
      responseCluster,
    ],
  };

  fetchMock.post({
    url: 'http://localhost/prpc/luci.analysis.v1.Clusters/BatchGet',
    body: request,
  }, {
    headers: {
      'X-Prpc-Grpc-Code': '0',
    },
    body: ')]}\'' + JSON.stringify(response),
  }, { overwriteRoutes: true });
};

export const mockQueryClusterFailures = (parent: string, failures: DistinctClusterFailure[] | undefined) => {
  const request: QueryClusterFailuresRequest = {
    parent: parent,
  };
  const response: QueryClusterFailuresResponse = {
    failures: failures,
  };
  fetchMock.post({
    url: 'http://localhost/prpc/luci.analysis.v1.Clusters/QueryClusterFailures',
    body: request,
  }, {
    headers: {
      'X-Prpc-Grpc-Code': '0',
    },
    body: ')]}\'' + JSON.stringify(response),
  }, { overwriteRoutes: true });
};

export const mockQueryExoneratedTestVariants = (parent: string, testVariants: ClusterExoneratedTestVariant[]) => {
  const request: QueryClusterExoneratedTestVariantsRequest = {
    parent: parent,
  };
  const response: QueryClusterExoneratedTestVariantsResponse = {
    testVariants: testVariants,
  };
  fetchMock.post({
    url: 'http://localhost/prpc/luci.analysis.v1.Clusters/QueryExoneratedTestVariants',
    body: request,
  }, {
    headers: {
      'X-Prpc-Grpc-Code': '0',
    },
    body: ')]}\'' + JSON.stringify(response),
  }, { overwriteRoutes: true });
};
