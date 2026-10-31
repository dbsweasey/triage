# triage

A simple LLM-based triage system built in Go with a TS + React frontend

### TODO
- updatestatus() in both stores and interface
- add t.Run test
- return ErrNotFound if ExecContext returns res.RowsEffected() == 0
- keep valid status in a list. return 400 for anything not in list
