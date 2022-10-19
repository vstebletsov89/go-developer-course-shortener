# go test -bench . -benchmem -benchtime 10s -memprofile base.pprof   (used this command to gather memory stats)
# go test -bench . -benchmem -benchtime 10s -memprofile result.pprof (used this command to compare results)
# go tool pprof -http=":8081" -diff_base base.pprof result.pprof