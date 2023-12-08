* Get user name to session
* proper auth checking
* tcp messaging layer
* startup datastore
* candidate datastore
* locking
* client
* session closing/+cleanup
* limit message handling to single threaded
  * "The managed device MUST send responses only in the order the requests were received." (pipelining)

* copy over rpc attrs to rpc-reply
* rpc-error - service layer errrs
  * error-type, error-tag, error-severity
  * error-app-tag, error-path
  * error-message, error-info

* filters (a lot of details, unclear how much work)
   
module x {
      namespace y;
}

   filter
      no filter 
                    select all from all modules

      empty filter 
                    selects nothing

      <x xmlns="y"> 
                    selects all data from module w/ns = "y"

      <containment xmlns="y">
         <selection />     
                    selects all <selection> nodes from module w/ns = "y"

      <containment xmlns="y">
         <selection>
            <content-matching1>xxx</content-matching1>
            <content-matching2>yyy</content-matching2>

                    selects only the <selection> nodes from module
                    with content-matching1 == xxx AND
                    content-matching2 == yyy 

      <containment xmlns="y">
         <selection1>
            <content-matching1>xxx</content-matching1>
            <content-matching2>yyy</content-matching2>
            <selection2>

                    selects only the <selection1> nodes from module
                    with content-matching1 == xxx AND
                    content-matching2 == yyy    
                    and ONLY elements:        
                       content-matching1
                       content-matching2
                       selection2          

      <containment>
         <containment>
           <selection1 />
           <selection2 />           

                    selects all <selection1> and <selection2> nodes from module


* edit-config
    default is merge

    test-option
      test-then-set  default : seems like setting all on buffer first, then on 
                               set again on final if no errors above
      set
      test-only

    error-option
      stop-on-error    default
      continue-on-error
      rolllback-on-error