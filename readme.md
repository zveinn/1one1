# Lynx

# TODO 
## refactoring
### add logger !!
## client UI
- 1. put UI on a different port
- 2. add a config for the UI struct
- 3. design handshake for the UI
    - this means making a golang UI client to test. 
- 4. .. see more info in excel sheet.

## controller data queries.. 
1. look into influx DB for backend? 
 - what does prometheus use for it's data source? 
2. Think about how you can make a query system that spans all servers...

# server queries.. 
## if we use data on disk.. 
- Each controller is essentially a database for a single cluster of machines, it does not go away or get deleted. How the client backs up his data is entirely up to him. 
- Each UI can connect to one or more clusters and combine them in a chart or toggle between them. 

# to check
- how many collectors can a controller handle. 
- How much bandwidth is each collector using?
     - check different update rates
     