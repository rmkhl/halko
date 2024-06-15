# Executor

## Storage (filesystem)

 halko/programs/... named programs
 halko/executed/... executed programs and their results
 halko/running/... currently running program and its result

## API

POST api/v1/programs (upload  program)
PUT api/v1/programs/name (replace program)
DELETE api/v1/programs/name

POST api/v1/running (program name -> start program)
DELETE api/v1/running (abort)
GET api/v1/running

## Logic

step 0: preheat (full fan, no humidifier)
    if preheat temperature is not given only preheat until 
    we reached the lowest delta of first step

## Program verification automatic adaptation

Last step minimum cannot be above oven starting temperature, if when program is started the oven is above the minimum, show warning

## Possible issues

Do we need ambient temperature
