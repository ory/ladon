<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Contributing to pq](#contributing-to-pq)
  - [Patch review](#patch-review)
  - [Bug reports](#bug-reports)
  - [New functionality](#new-functionality)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Contributing to pq

`pq` has a backlog of pull requests, but contributions are still very
much welcome. You can help with patch review, submitting bug reports,
or adding new functionality. There is no formal style guide, but
please conform to the style of existing code and general Go formatting
conventions when submitting patches.

### Patch review

Help review existing open pull requests by commenting on the code or
proposed functionality.

### Bug reports

We appreciate any bug reports, but especially ones with self-contained
(doesn't depend on code outside of pq), minimal (can't be simplified
further) test cases. It's especially helpful if you can submit a pull
request with just the failing test case (you'll probably want to
pattern it after the tests in
[conn_test.go](https://github.com/lib/pq/blob/master/conn_test.go).

### New functionality

There are a number of pending patches for new functionality, so
additional feature patches will take a while to merge. Still, patches
are generally reviewed based on usefulness and complexity in addition
to time-in-queue, so if you have a knockout idea, take a shot. Feel
free to open an issue discussion your proposed patch beforehand.
