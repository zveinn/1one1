# Lynx

## places to consider recovering.

## todo
 - save data to file
 - pipe data to infrastructure

## infra notes

### 1. step - remove server from cloudflare if overutilized
 - check if limits have been breached for more then 10 seconds
 - if so then remove server from CF
 - if stats stabilize on server we can add it back to CF
 ### 2. step - enabled backup server
 - we enable a backup server once we only have X server left on the active list
 - if we have enabled a backup server we spawn a new machine to take it's place? - rethink
 ### 3. step - spawn new servers if needed
- we only spawn new servers as a last resort
