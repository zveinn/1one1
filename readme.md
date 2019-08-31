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



# NEW LYNX!

1. normalize everything.

# Datas
1. host data
 - Check for changes every X seconds
2. Normalized
 - Just order the bytes and have them in a static queue
 - byte/byte/byte/byte/byte/byte
 - Cutom normalized data:
 - controlByte/byte/byte/byte/byte/byte
 - Get customized: controlByte/index
3. NETWORK?
 - https://superuser.com/questions/356907/how-to-get-real-time-network-statistics-in-linux-with-kb-mb-bytes-format-and-for
 - use this to calculate bytes.

# Data packet formats
1. controlByte:

# Todo
1. remake the Excel sheet so that is has some real structure
2. revisit all data collection
3. Document all data transmission protocols
4. add stats that are needed for new protocols. 
5. Implenent proper logging and error handling.
6. ....