# Agent Guidelines for go-go-golems go projects

## Build Commands
- Run a binary in XXX/YYY/FOOO: `go run ./XXX/YYY/FOOO` . Use this instead of build + ./XXX.
- Build: `go build ./...`
- Test: `go test ./...`
- Run single test: `go test ./pkg/path/to/package -run TestName`
- Generate: `go generate ./...`
- Lint: `golangci-lint run -v` or `make lint`
- Format: `go fmt ./...`

IMPORTANT: To run a server and do some interaction with it, use tmux, this makes it very easy to kill a server. 
Use capture-pane to read the output.

## Project Structure
- `cmd/`: CLI commands and entry points
- `pkg/`: Library code organized by domain
- `examples/`: Example configurations and usage
- `doc/`: Documentation
- `ttmp/YYYY-MM-DD/`: this is where all temporary documentation as well as debugging logs and other reports go

<runningProcessesGuidelines>
- When testing TUIs, use tmux and capture-pane to interact with the UI.
- When using tmux, try to batch as many commands as possible when using send-keys.
- When running long-running processes (servers, etc...), use tmux to more easily interact and kill them.
- Kill a process using port $PORT: `lsof-who -p $PORT -k`. When building a web server, ALWAYS use this command to kill the process.
</runningProcessesGuidelines>

<goGuidelines>
- When implementing go interfaces, use the var _ Interface = &Foo{} to make sure the interface is always implemented correctly.
- When building web applications, use bootstrap CSS unless otherwise indicated.
- Always use a context argument when appropriate.
- Use cobra for command-line applications.
- Use the "defaults" package name, instead of "default" package name, as it's reserved in go.
- Use github.com/pkg/errors for wrapping errors.
- When starting goroutines, use errgroup.

- Only use the toplevel go.mod, don't create new ones.
- When writing a new experiment / app, add zerolog logging to help debug and figure out how it works, add --log-level flag to set the log level.
- When using go:embed, import embed as `_ "embed"`
- When using build tagged features, make sure the software compiles without the tag as well
</goGuidelines>

<libraryGuidelines>
- when interfacing with the google gemini/genai APIs, use the new https://pkg.go.dev/google.golang.org/genai package
</libraryGuidelines>

<webGuidelines>
- Use bun, react and rtk-query. Use typescript.
- Use bootstrap for styling.
- Store css, html and js in different files in a static directory.
- Use go:embed to serve static files.
- Use templ for go templates, assume I'm running templ generate -watch in the background.
- Always serve static files under /static/ URL paths, never directly under functional paths like /admin/
</webGuidelines>

<debuggingGuidelines>
If me or you the LLM agent seem to go down too deep in a debugging/fixing rabbit hole in our conversations, remind me to take a breath and think about the bigger picture instead of hacking away. Say: "I think I'm stuck, let's TOUCH GRASS".  IMPORTANT: Don't try to fix errors by yourself more than twice in a row. Then STOP. Don't do anything else.

</debuggingGuidelines>

<generalGuidelines>
Don't add backwards compatibility layers or adapters unless explicitly asked. If you think there is a need for a backwards compatibility or adapting to an existing interface, STOP AND ASK ME IF THAT IS NECESSARY. Usually, I don't need backwards compatibility.

If it looks like your edits aren't applied, stop immediately and say "STOPPING BECAUSE EDITING ISN'T WORKING".

Run the format_file tool at the end of each response.
</generalGuidelines>

<gitGuidelines>
After each successful run, output a git commit message as yaml, using gitmoji for th title.
If new APIs, endpoints or CLI commands were added, document them in the commit message.
Make a list of the tests that were run and their results.

Avoid general statements like "created extensible foundation" or "this allows for...".

title: <title>
description: <description>

and store in directory root as .git-commit-message.yaml
</gitGuidelines>
