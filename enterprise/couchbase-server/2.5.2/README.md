
This README will guide you through running Couchbase Server under Docker.

# Background Information

## Networking

## Volumes

# Common Deployment Scenarios

## Single container on single host (easy)

                                                                                      
       ┌───────────────────────┐                                                      
       │   Host OS (Ubuntu)    │                                                      
       │  ┌─────────────────┐  │                                                      
       │  │  Container OS   │  │                                                      
       │  │    (CentOS)     │  │                                                      
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
       │   Host OS (Ubuntu)    │  │   Host OS (Ubuntu)    │  │   Host OS (Ubuntu)    │
       │  ┌─────────────────┐  │  │  ┌─────────────────┐  │  │  ┌─────────────────┐  │
       │  │  Container OS   │  │  │  │  Container OS   │  │  │  │  Container OS   │  │
       │  │    (CentOS)     │  │  │  │    (CentOS)     │  │  │  │    (CentOS)     │  │
       │  │  ┌───────────┐  │  │  │  │  ┌───────────┐  │  │  │  │  ┌───────────┐  │  │
       │  │  │ Couchbase │  │  │  │  │  │ Couchbase │  │  │  │  │  │ Couchbase │  │  │
       │  │  │  Server   │  │  │  │  │  │  Server   │  │  │  │  │  │  Server   │  │  │
       │  │  └───────────┘  │  │  │  │  └───────────┘  │  │  │  │  └───────────┘  │  │
       │  │                 │  │  │  │                 │  │  │  │                 │  │
       │  └─────────────────┘  │  │  └─────────────────┘  │  │  └─────────────────┘  │
       │                       │  │                       │  │                       │
       └───────────────────────┘  └───────────────────────┘  └───────────────────────┘


## Running in environments with SDN (easy)

Some cloud providers, such as Amazon ECS and Joyent's Triton Container Cloud, provide Software Defined Networking (SDN) which simplifies the networking setup required to run Couchbase Server.

                                                                                      
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

                                                                                      
       ┌──────────────────────────────────────────────────────────┐                   
       │                     Host OS (Ubuntu)                     │                   
       │                                                          │                   
       │  ┌───────────────┐ ┌───────────────┐  ┌───────────────┐  │                   
       │  │ Container OS  │ │ Container OS  │  │ Container OS  │  │                   
       │  │   (CentOS)    │ │   (CentOS)    │  │   (CentOS)    │  │                   
       │  │ ┌───────────┐ │ │ ┌───────────┐ │  │ ┌───────────┐ │  │                   
       │  │ │ Couchbase │ │ │ │ Couchbase │ │  │ │ Couchbase │ │  │                   
       │  │ │  Server   │ │ │ │  Server   │ │  │ │  Server   │ │  │                   
       │  │ └───────────┘ │ │ └───────────┘ │  │ └───────────┘ │  │                   
       │  │               │ │               │  │               │  │                   
       │  └───────────────┘ └───────────────┘  └───────────────┘  │                   
       │                                                          │                   
       └──────────────────────────────────────────────────────────┘                   


