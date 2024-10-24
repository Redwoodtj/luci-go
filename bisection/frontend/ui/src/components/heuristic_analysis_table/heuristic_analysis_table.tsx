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


import './heuristic_analysis_table.css';

import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';

import { HeuristicAnalysisTableRow } from './heuristic_analysis_table_row/heuristic_analysis_table_row';
import { NoDataMessageRow } from '../no_data_message_row/no_data_message_row';
import {
  HeuristicAnalysisResult,
  HeuristicSuspect,
  isAnalysisComplete,
} from '../../services/luci_bisection';

interface Props {
  result?: HeuristicAnalysisResult;
}

function getInProgressRow() {
  return (
    <NoDataMessageRow message='Heuristic analysis is in progress' columns={4} />
  );
}

function getRows(suspects: HeuristicSuspect[] | undefined) {
  if (!suspects || suspects.length === 0) {
    return <NoDataMessageRow message='No suspects to display' columns={4} />;
  } else {
    return suspects.map((suspect) => (
      <HeuristicAnalysisTableRow
        key={suspect.gitilesCommit.id}
        suspect={suspect}
      />
    ));
  }
}

export const HeuristicAnalysisTable = ({ result }: Props) => {
  return (
    <TableContainer component={Paper} className='heuristic-table-container'>
      <Table className='heuristic-table' size='small'>
        <TableHead>
          <TableRow>
            <TableCell>Suspect CL</TableCell>
            <TableCell>Confidence</TableCell>
            <TableCell>Score</TableCell>
            <TableCell>Justification</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {result && isAnalysisComplete(result.status)
            ? getRows(result.suspects)
            : getInProgressRow()}
        </TableBody>
      </Table>
    </TableContainer>
  );
};
