# factomd-sample-event-consumer
An example project where events from the factomd live feed api is dumped into the console.  
You can enable the live feed api by settings EnableLiveFeedAPI to true in the factomd.conf file.   

[LiveFeedAPI]  
EnableLiveFeedAPI                     = false  
EventReceiverProtocol                 = tcp  
EventReceiverAddress                  = 127.0.0.1  
EventReceiverPort                     = 8040  
EventFormat                           = protobuf  
MuteReplayDuringStartup               = false  
ResendRegistrationsOnStateChange      = false  
ContentFilterMode                     = SendOnRegistration   
  
As of writing the live feed api functionality is only available in fork branch  
https://github.com/bi-foundation/factomd/tree/FD-1150_live_api  

