// FactorLog is a logging infrastructure for Go that provides numerous
// logging functions for whatever your style may be. It could easily
// be a replacement for Go's log in the standard library (though it
// doesn't support functions such as `SetFlags()`).
//
// Basic usage:
//   import log "github.com/kdar/factorlog"
//   log.Print("Hello there!")
//
// Setting your own format:
//   import os
//   import "github.com/kdar/factorlog"
//   log := factorlog.New(os.Stdout, "%T %f:%s %M"
//   log.Print("Hello there!")
//
// Setting the verbosity and testing against it:
//   import os
//   import "github.com/kdar/factorlog"
//   log := factorlog.New(os.Stdout, "%T %f:%s %M"
//   log.SetVerbosity(2)
//   log.V(1).Print("Will print")
//   log.V(3).Print("Will not print")
//
// If you care about performance, you can test for verbosity this way:
//   if log.IsV(1) {
//     log.Print("Hello there!")
//   }
//
// Format verbs:
//   %T - Time: 15:04:05.000000
//   %t - Time: 15:04:05
//   %D - Date: 2006-01-02
//   %d - Date: 2006/01/02
//   %L - Severity
//   %l - Short severity
//   %F - File name (full path)
//   %f - Short file name
//   %x - Extra short file name (no go suffix)
//   %s - Source line number
//   %M - Message
//   %P - Package path and function (e.g. sql.New)
//   %p - Function name
//   %% - Percent sign
package factorlog
