## TODO
Roughly in order of importance

* tcp messaging layer
* startup datastore
* candidate datastore
* client
* adjust capabilties to properly reflect
* locking
* expand on basic xpath support
  - relative paths
  - hardening parser
  - list item selection
* edit-config special options
   - test-option
      test-then-set  default : seems like setting all on buffer first, then on 
                               set again on final if no errors above
      set
      test-only
    error-option
      stop-on-error    default
      continue-on-error
      rollback-on-error
* limit message handling to single threaded
  * "The managed device MUST send responses only in the order the requests were received." (pipelining)
* enable python
* update docs

* rpc-error - service layer errrs
  * error-type, error-tag, error-severity
  * error-app-tag, error-path
  * error-message, error-info

## Done
* Get user name to session (done: bare minimum)
* proper auth checking
