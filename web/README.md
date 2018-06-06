# Promviz-front [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)
[nghialv/promviz](https://github.com/nghialv/promviz) web front-end application  

This fork of Netflix's [vizceral-example](https://github.com/Netflix/vizceral-example) contains these new features:
* Replaying
* Connection Chart
* Node Coloring
* Class Filtering
* Notice Filtering

...

# Install
This application is one of promviz's components.  
To install, please refer to [nghialv/promviz#install](https://github.com/nghialv/promviz#install).  

# Install & Run Independently
```
npm install
npm run dev
```
or, you can use Docker to run:
```
docker build -t <name>/promviz-front .
docker run -p 8080:8080 -d <name>/promviz-front
```
then, you can view the top page at http://localhost:8080/

### Public Docker Repository
[mjhddevlion/promviz-front](https://hub.docker.com/r/mjhddevlion/promviz-front/)

# Configuration
There are 2 ways to configure this application:  
1. edit .env file
1. set environment variables

You can customize this application's behavior with these variables:  
```
UPDATE_URL: endpoint of promviz server  
INTERVAL: interval between fetches (ms)  
MAX_REPLAY_OFFSET: limit of replaying offset (s)  
```

# Contributing
Welcome PRs!
