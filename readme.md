# Lynx

# TODO 
### We now have a current stats buffer
- send that buffer on user connect
- send also static data
### Make static data buffer
### finish static data.
### ALL DISKS!
## refactoring

### add logger !!


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
     


# UI stuff
- handshake 
     - To controller: config
     - From controller: datapoints



# Data point conversions
1. from collector: index/index/value = type index/position index/ value
2. at controller for UI: index.index


# Size
-? 
# luminocity
- ?



# alarms
TODO: 
1. Add to disk
2. ????



