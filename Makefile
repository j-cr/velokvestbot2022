GO = go
OUT = ./out
TEST_SUFFIX = printf "\nPASSED\n\n" || echo

run :
	$(GO) run me/velokvestbot/cmd/bot

build :
	$(GO) build -o $(OUT) me/velokvestbot/cmd/bot

test :
	$(GO) test -v ./... && $(TEST_SUFFIX)

clean:
	rm -f $(OUT)/*
