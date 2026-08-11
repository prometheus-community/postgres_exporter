// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import "github.com/prometheus-community/postgres_exporter/config"

const (
	buffercacheSummarySubsystem      = config.CollectorBuffercacheSummary
	databaseSubsystem                = config.CollectorDatabase
	databaseWraparoundSubsystem      = config.CollectorDatabaseWraparound
	locksSubsystem                   = config.CollectorLocks
	longRunningTransactionsSubsystem = config.CollectorLongRunningTransactions
	postmasterSubsystem              = config.CollectorPostmaster
	processIdleSubsystem             = config.CollectorProcessIdle
	replicationSubsystem             = config.CollectorReplication
	replicationSlotsSubsystem        = config.CollectorReplicationSlots
	rolesSubsystem                   = config.CollectorRoles
	settingsSubsystem                = config.CollectorSettings
	statActivitySubsystem            = config.CollectorStatActivity
	statActivityAutovacuumSubsystem  = config.CollectorStatActivityAutovacuum
	statArchiverSubsystem            = config.CollectorStatArchiver
	bgWriterSubsystem                = config.CollectorStatBGWriter
	statCheckpointerSubsystem        = config.CollectorStatCheckpointer
	statDatabaseSubsystem            = config.CollectorStatDatabase
	progressVacuumSubsystem          = config.CollectorStatProgressVacuum
	statReplicationSubsystem         = config.CollectorStatReplication
	statStatementsSubsystem          = config.CollectorStatStatements
	userTableSubsystem               = config.CollectorStatUserTables
	statWalReceiverSubsystem         = config.CollectorStatWalReceiver
	statioUserIndexesSubsystem       = config.CollectorStatioUserIndexes
	statioUserTableSubsystem         = config.CollectorStatioUserTables
	walSubsystem                     = config.CollectorWal
	xlogLocationSubsystem            = config.CollectorXlogLocation
)
