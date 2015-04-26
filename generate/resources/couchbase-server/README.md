
This README will guide you through running Couchbase Server under Docker.

# Background Information

## Networking

## Volumes

# Common Deployment Scenarios

## Single container on single host (easy)
                                                  
                                                                                      
       ┌───────────────────────┐                                                      
       │ Host OS (eg, Ubuntu)  │                                                      
       │  ┌─────────────────┐  │                                                      
       │  │Docker Container │  │                                                      
       │  │     Engine      │  │                                                      
       │  │  ┌───────────┐  │  │                                                      
       │  │  │ Couchbase │  │  │                                                      
       │  │  │  Server   │  │  │                                                      
       │  │  └───────────┘  │  │                                                      
       │  │                 │  │                                                      
       │  └─────────────────┘  │                                                      
       │                       │                                                      
       └───────────────────────┘                                                      
                                                                                      

## Multiple hosts, single container on each host (easy)

                                                                                      
       ┌───────────────────────┐  ┌───────────────────────┐  ┌───────────────────────┐
       │ Host OS (eg, Ubuntu)  │  │ Host OS (eg, Ubuntu)  │  │ Host OS (eg, Ubuntu)  │
       │  ┌─────────────────┐  │  │  ┌─────────────────┐  │  │  ┌─────────────────┐  │
       │  │Docker Container │  │  │  │Docker Container │  │  │  │Docker Container │  │
       │  │     Engine      │  │  │  │     Engine      │  │  │  │     Engine      │  │
       │  │  ┌───────────┐  │  │  │  │  ┌───────────┐  │  │  │  │  ┌───────────┐  │  │
       │  │  │ Couchbase │  │  │  │  │  │ Couchbase │  │  │  │  │  │ Couchbase │  │  │
       │  │  │  Server   │  │  │  │  │  │  Server   │  │  │  │  │  │  Server   │  │  │
       │  │  └───────────┘  │  │  │  │  └───────────┘  │  │  │  │  └───────────┘  │  │
       │  │                 │  │  │  │                 │  │  │  │                 │  │
       │  └─────────────────┘  │  │  └─────────────────┘  │  │  └─────────────────┘  │
       │                       │  │                       │  │                       │
       └───────────────────────┘  └───────────────────────┘  └───────────────────────┘


## Running in environments with SDN (easy)

Some cloud providers, such as Joyent, provide Software Defined Networking (SDN) which simplifies the networking setup required to run Couchbase Server.

                                                                                      
       ┌─────────────────────────────────────────────────────┐                        
       │                     Environment                     │                        
       │                                                     │                        
       │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │                        
       │  │  Couchbase  │  │  Couchbase  │  │  Couchbase  │  │                        
       │  │   Server    │  │   Server    │  │   Server    │  │                        
       │  │             │  │             │  │             │  │                        
       │  │ private ip: │  │ private ip: │  │ private ip: │  │                        
       │  │ 10.20.21.1  │  │ 10.20.21.1  │  │ 10.20.21.1  │  │                        
       │  │             │  │             │  │             │  │                        
       │  │ public ip:  │  │ public ip:  │  │ public ip:  │  │                        
       │  │ 62.87.22.8  │  │ 62.87.22.8  │  │ 62.87.22.8  │  │                        
       │  └─────────────┘  └─────────────┘  └─────────────┘  │                        
       └─────────────────────────────────────────────────────┘
       

## Multiple containers per host(s) (hard)

                                                                                      
       ┌──────────────────────────────────────────────────┐                           
       │               Host OS (eg, Ubuntu)               │                           
       │  ┌─────────────────────────────────────────────┐ │                           
       │  │           Docker Container Engine           │ │                           
       │  │                                             │ │                           
       │  │  ┌───────────┐  ┌───────────┐ ┌───────────┐ │ │                           
       │  │  │ Couchbase │  │ Couchbase │ │ Couchbase │ │ │                           
       │  │  │  Server   │  │  Server   │ │  Server   │ │ │                           
       │  │  └───────────┘  └───────────┘ └───────────┘ │ │                           
       │  │                                             │ │                           
       │  └─────────────────────────────────────────────┘ │                           
       │                                                  │                           
       └──────────────────────────────────────────────────┘                           

