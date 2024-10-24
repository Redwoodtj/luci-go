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

import dayjs from 'dayjs';

import Chip from '@mui/material/Chip';
import CloseIcon from '@mui/icons-material/Close';
import DoneIcon from '@mui/icons-material/Done';
import Typography from '@mui/material/Typography';
import Link from '@mui/material/Link';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';

import {
  ExoneratedTestVariant,
} from '@/components/cluster/cluster_analysis_section/exonerations_tab/model/model';
import {
  invocationName,
  failureLink,
} from '@/tools/urlHandling/links';
import {
  TestVariantFailureRateAnalysisVerdictExample,
} from '@/services/test_variants';
import CLList from '@/components/cl_list/cl_list';

interface Props {
  testVariant: ExoneratedTestVariant;
}

const FlakyCriteriaSection = ({
  testVariant,
}: Props) => {
  return (
    <>
      <Typography variant="h6">
        Purpose
      </Typography>
      <Typography paragraph>
        Exonerates test variants that are so flaky that (inexpensive) retries are no longer effective at mitigating failures.
      </Typography>
      <Typography variant="h6">
        Definition
      </Typography>
      <Typography component='div' paragraph>
        <Chip
          variant='outlined'
          color={testVariant.runFlakyVerdicts1wd >= 1 ? 'success' : 'default'}
          icon={testVariant.runFlakyVerdicts1wd >= 1 ? (<DoneIcon/>) : (<CloseIcon/>)}
          label={
            <>Run-flaky verdicts in the last weekday <strong data-testid='flaky_verdicts_1wd'>(current value: {testVariant.runFlakyVerdicts1wd})</strong> &gt;= 1</>
          }
        />&nbsp;AND&nbsp;
        <Chip
          variant='outlined'
          color={testVariant.runFlakyVerdicts5wd >= 3 ? 'success' : 'default'}
          icon={testVariant.runFlakyVerdicts5wd >= 3 ? (<DoneIcon/>) : (<CloseIcon/>)}
          label={
            <>Run-flaky verdicts in the last five weekdays <strong data-testid='flaky_verdicts_5wd'>(current value: {testVariant.runFlakyVerdicts5wd})</strong> &gt;= 3</>
          }
        />&nbsp;.
      </Typography>
      <Typography component='div'>
        Where:
        <ul>
          <li>
            <strong>Run-flaky verdicts</strong>: Verdict which required a step-level (i.e. swarming task-level) retry
            to obtain an expected result, filtered to at most one verdict per distinct changelist. These are verdicts
            for which inexpensive (in-step) retries were not effective.
          </li>
          <li>
            <strong>Weekday</strong>: An interval of at least 24 hours, which starts at a given time on a calendar weekday
            and ends at the same time on the following calendar weekday. Includes the intervening weekend (if any).
          </li>
        </ul>
      </Typography>
      <Typography variant="h6">
        Recent run-flaky verdicts
      </Typography>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>Invocation</TableCell>
            <TableCell>Changelist and patchset</TableCell>
            <TableCell>Timestamp</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {testVariant.runFlakyVerdictExamples.map((verdict :TestVariantFailureRateAnalysisVerdictExample, i: number) => {
            return (
              <TableRow key={i.toString()}>
                <TableCell>
                  <Link
                    aria-label="invocation id"
                    sx={{ mr: 2 }}
                    href={failureLink(verdict.ingestedInvocationId, testVariant.testId)}
                    target="_blank"
                  >
                    {invocationName(verdict.ingestedInvocationId)}
                  </Link>
                </TableCell>
                <TableCell>
                  <CLList changelists={verdict.changelists || []}></CLList>
                </TableCell>
                <TableCell>
                  {dayjs(verdict.partitionTime).fromNow()}
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </>
  );
};

export default FlakyCriteriaSection;
