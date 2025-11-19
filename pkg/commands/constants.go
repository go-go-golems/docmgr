package commands

import "strings"

// Column name constants used across docmgr list commands
const (
	ColTicket      = "ticket"
	ColTitle       = "title"
	ColStatus      = "status"
	ColTopics      = "topics"
	ColPath        = "path"
	ColLastUpdated = "last_updated"
	ColTasksOpen   = "tasks_open"
	ColTasksDone   = "tasks_done"

	ColDocType = "doc_type"

	ColIndex   = "index"
	ColChecked = "checked"
	ColText    = "text"
	ColFile    = "file"

	ColCategory    = "category"
	ColSlug        = "slug"
	ColDescription = "description"
)

var ColumnsTickets = []string{ColTicket, ColTitle, ColStatus, ColTopics, ColTasksOpen, ColTasksDone, ColPath, ColLastUpdated}
var ColumnsDocs = []string{ColTicket, ColDocType, ColTitle, ColStatus, ColTopics, ColPath, ColLastUpdated}
var ColumnsTasksList = []string{ColIndex, ColChecked, ColText, ColFile}
var ColumnsVocabList = []string{ColCategory, ColSlug, ColDescription}

func ColumnsListString(cols []string) string { return strings.Join(cols, ",") }
