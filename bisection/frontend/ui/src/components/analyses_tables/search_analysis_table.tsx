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

import './analyses_table.css';

import { useQuery } from 'react-query';

import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';

import { AnalysisTableRow } from './analysis_table_row/analysis_table_row';
import { NoDataMessageRow } from '../no_data_message_row/no_data_message_row';
import {
  getLUCIBisectionService,
  QueryAnalysisRequest,
} from '../../services/luci_bisection';

interface Props {
  bbid: string | null | undefined;
}

export const SearchAnalysisTable = ({ bbid }: Props) => {
  const bisectionService = getLUCIBisectionService();

  const {
    isLoading,
    isError,
    isSuccess,
    data: response,
    error,
  } = useQuery(
    ['analysis', bbid],
    async () => {
      const request: QueryAnalysisRequest = {
        buildFailure: {
          bbid: bbid!,
          // TODO: update this once other failure types are analyzed
          failedStepName: 'compile',
        },
      };

      return await bisectionService.queryAnalysis(request);
    },
    {
      // only use the query if a Buildbucket ID has been provided
      enabled: !!bbid,
    }
  );

  if (isLoading) {
    return (
      <Box display='flex' justifyContent='center' alignItems='center'>
        <CircularProgress />
      </Box>
    );
  }

  if (isError) {
    return (
      <div className='section'>
        <Alert severity='error'>
          <AlertTitle>Issue searching by build</AlertTitle>
          {/* TODO: display more error detail for input issues e.g.
                  Build not found, No analysis for that build, etc */}
          An error occurred when searching for analysis using Buildbucket ID "
          {bbid}":
          <Box sx={{ padding: '1rem' }}>{`${error}`}</Box>
        </Alert>
      </div>
    );
  }

  let analysis = null;
  let buildIsFirstFailed = false;
  if (
    isSuccess &&
    response &&
    response.analyses &&
    response.analyses.length > 0
  ) {
    analysis = response.analyses[0];
    buildIsFirstFailed = analysis.firstFailedBbid === bbid;
  }

  return (
    <>
      <TableContainer className='analyses-table-container' component={Paper}>
        <Table className='analyses-table' size='small'>
          <TableHead>
            <TableRow>
              <TableCell>Buildbucket ID</TableCell>
              <TableCell>Created time</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Failure type</TableCell>
              <TableCell>Duration</TableCell>
              <TableCell>Builder</TableCell>
              <TableCell>Culprit CL</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {analysis ? (
              <AnalysisTableRow analysis={analysis} />
            ) : (
              <NoDataMessageRow
                message={`No analysis found for build ${bbid}`}
                columns={7}
              />
            )}
          </TableBody>
        </Table>
      </TableContainer>
      {analysis && !buildIsFirstFailed && (
        <div className='section'>
          <Alert severity='info'>
            <AlertTitle>Found related analysis</AlertTitle>
            The above analysis is related to build {bbid}; there is an earlier
            failed build associated with it.
          </Alert>
        </div>
      )}
    </>
  );
};
