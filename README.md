# triage

A simple LLM-based triage system built in Go with a TS + React frontend

### TODO

- build get handler
- wire to main!
- test 500 errors (main_test.go)
- test get handler
- test create handler
    - try create triage on non-existant ticket
    - create triage on ticket
    - invalid id
    - follow other error paths
