An HTTP middleware that will recover from any panics in an application, even if the responseWriter has been partially written to, and then output the stack trace if the application is in development mode.
The stack trace contains navigation links to the source code files that are rendered with syntax highlighting for easy debugging.
