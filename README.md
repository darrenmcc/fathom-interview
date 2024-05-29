# fathom-interview

Any questions that you would want/need answered by the product manager (or team lead) before starting work on the behavior
  - Do we have an upper bound on meeting length? 
  - Do we need any authentication mechanism for this service? 
  - Do we have any SLO/SLAs regarding uptime, latency, connections or throughput?
  - Is it possible to predict traffic patterns in the near term? 
  - Do we have any budgets on server/storage costs for running this service?
Any assumptions that you would be comfortable making (rather than addressing them to the PM or lead) in the course of performing this task
  - I assume the service should seek to minimize downtime, latency and data loss between the client and the server. 
Any assumptions that you made for the purposes of this exercise that you would not make in the normal course of performing this task
  - I’m assuming K8s is configured with some sort of liveness probe that we can signal (via a healthcheck endpoint) that the service is ready to shut down.
  - I’m assuming we’re doing some sort of in-band processing of the stream and therefore can’t (if it’s at all possible) stream directly from the client into some sort of cloud storage. 
  - The stream session needs to be sticky and can’t be retried if the connection is temporarily lost. 
  - We don’t need any sort of authn/authz mechanism.
  - We don’t need any monitoring besides basic logging.
Any UI/UX concerns that would affect how you implement the behavior
  - How does the service handle spotty connections from the client? Can we retry temporarily lost connections, or is the stream simply dead at that point? Can we link two streams together to create one summary/transcription? 
Any architectural/system/performance concerns that would affect how you implement the behavior
  - There is no timeout on individual connections, so in theory a meeting could go on indefinitely and keep the connection open even after the server marks itself for shutdown. 
