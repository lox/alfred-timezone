SOURCES := $(wildcard *.go)
BIN := alfred-timezone
FILES := $(BIN) info.plist images/
WORKFLOW := Timezone.alfredworkflow

all: $(BIN) $(WORKFLOW)

$(WORKFLOW): $(FILES)
	zip -j "$@" $^

$(BIN): $(SOURCES)
	go build -o $(BIN) $(SOURCES)

clean:
	rm $(BIN) $(WORKFLOW)