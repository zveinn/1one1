# lynx

# working on now....
1. Klára static

# Notes
1. namesapace format: NS|namespace:(100|ms),namespace:(1|m),namespace:(1|h),namespace:(1|d),namespace:(x)
      - default collection: 1 second: ( maybe lower ???)
      - h:hour
      - m:minute
      - ms: millisecond ( minimum 10 ms ?? or just let people decide ? )
      - x: disable namespace collection all together
      - fastest delivery time is 1 second: ( maybe less? ) .. if you data point is being collected slower it will be delivered according to it's collection time ( on the next second tick )
2. 

#collecting
### order 
 - memory,load,disk,network
-- stats.go
1. CollectDynamicData
2. PrepareDataForShipping

# data format
 - timestamp:slot,slot,slot,slot,slot,slot:slot,slot,slot,slot,....
 - , = stat seperator
 - : = stat type seperator




 # todo
- diff on static stats