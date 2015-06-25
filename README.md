## An Indexer Daemon for the Go Programming Language

Gosemki is command line / daemon tool intended to be integrated with source code editor or IDE. Using this tool with JSON API editor can get
- highlight hints for all identifiers in code
- syntax and semantic errors
- code folding hints

Gosemki uses client/server architecture to implement caching in future releases.

### JSON format
Results in JSON format use following scheme:
```
{ "Ranges": [{  // List of hints for identifiers highlighting in editor
    "lin": 1,       // Line number where identifier placed
    "col": 2,       // Column where identifier starts
    "off": 2,       // Byte offset from source file start to first char of identifier
    "len": 4,       // Length of identifier
    "knd": "pkg"    // 'pkg' for imported packages, 'con' for constants, 'typ' for types, 'var' for variables, 'fun' for funcs, 'lbl' for goto labels and 'fld' for struct fields
  }],
  "Outline": [{  // List of items for document outline
    "lin": 1,       // Line number where identifier placed
    "col": 2,       // Column where identifier starts
    "off": 2,       // Byte offset from source file start to first char of identifier
    "str": 4,       // Title of outline item
    "knd": "pkg"    // 'pkg' for imported packages, 'con' for constants, 'typ' for types, 'var' for variables, 'fun' for funcs, 'lbl' for goto labels
  }],
  "Errors": [{  // List of syntax and semantic errors
      "lin": 1,     // Line number where error occured
      "col": 2,     // Column where error starts
      "offset": 2,  // Byte offset from source file start to the place of error
      "len": 4,     // Length of errorneous code
      "msg": "..."  // Error message from Go compiler
  }],
  "Folds": [{   // Lists of ranges for code folding in editor
      "from": 12    // First line of code folding range
      "to": 20      // Last line of code folding range
  }],
  InPanic: false // This flag is true after daemon panic occured
}
```
