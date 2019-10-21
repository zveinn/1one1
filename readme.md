# Lynx

# TODO 
### We now have a current stats buffer
- send that buffer on user connect
- send also static data
### Make static data buffer
### finish static data.
### ALL DISKS!
## refactoring



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






# NEW NOTES FOR FINAL SPRINT

# UP TO DATE FLOW
1. Start the brain
1.1 read configs
1.2 listen for controllers
2. start controller
2.1 controller looks for brain address in file or save to file.
2.2 connect to brain
 (from brain)0x0 = error, 0x1 = succcess, 0x2= brain config, 0x3= new address
2.3 receive configs
2.4 start go routines
3. start collector
3.1 connect to controller
3.1.1 send tag
3.1.2 receive namespaces
3.1.2 change active status

3.2 receive configs
3.3 collect stats and send to controller


# UI 
1. connect to brain
1.1 get controller list
1.2 connect to controllers
1.3 get collector list
1.4 start receiving stats

# UI NOTES
1. Grouping has to happen in the front-end
2. Can we somehow achive a similar layout with only css and html? that would be epic. 
     - maybe we can have a front and side view at the same time??
3. NETWORKING:
     - can we somehow determine link strength?
     - UI display: nuber on cube: 10.00 = GB .. 120mb = 0.12 ( above 1GB we ignore second decimals)
     - UI display: draw a line up and down for the network ? ... 
4. Possible alert types:
     - line from the left, top, right, bottom.. that goes outside the box
     - crosshairs
     - border highlight
     - luminocity(kind of like its heating...)
     - make a second bottom box under the current bottom layer that has sections indicating ??
          -
5. possible different display modes
      - reflected cube
      - thermometer next to cube showing network
      - thermometer only for all indexes ??
      - thermometer with segmented buttom for tag grouping... ?
      - cutom thermometer
      - sphere ?!
      - floor plan + items...

# stats index
1 = cpu used
2 = disk used / ( should I do all disks ... or maybe add a nr6+ where you can pick additional drivers? )
3 = memory used
4 = network in bytes per second
5 = network out bytes per second



# DATA ON DISK
1. rsyslog option on controller or maybe even collectors???.. that would be sick.
2. zip bulks on disk? or will that slo us down too much??
3. make it optional to save history.. options (off, disk, rsyslog, sql?, redis?..)

format:
ns(seperated by dots..)/year/month/day/hour/minute/datapoint