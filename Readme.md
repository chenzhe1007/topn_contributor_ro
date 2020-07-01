## Topn_Contributor_Ro

### Requirement 
the only requriement is to have docker installed

### Basic Setup
the docker env is good for both development and run time

#### to build the code
```
    ./enter
    go build
```

### to run the job
```
   ./${taskname} --help
```

### sample run
```
    ./zchen_topn_contributor_ro -owner awslabs -limit 20 -concurrency 5 -access-token <access token> 
```

### to log out of the container
```
    exit
```

The execution is broken up into three phases

Repolist Fetch (1 go routine)
--------------->[ contribute list job channel  ]       [   contribution counts channel ]
                                                                               
                         ^  ^                                  ^               ^
                         |  |                                  |               |
                                                               |         1 go routine lisening
                   contribute list processor  -----------------
                   a pool of go routine
                   lisenting  on this chanel                                  

Design Notes:
1. the contribute list processor will only spin up the specified concurrency and tops at 20
   (I read the best practice of github api, they don't recommand concurrent call at all, but
    since this is our requirement, so i tops it at 20)

2. the fetcher package is responsible to handle http related tasks
a. pagination is done using the LINK in the header
b. ratelimit is considered when the allowable request for the hour is less than 1000

3. The topn part is a minHeap with size of the total number of contributor intested
   space complexity is O(n) n is the total number of contributors in owner repo 
   time complexity to process is O(n(logk)) k is the size of the contributor we intereseted in

   One potential optizmiation of this step is to use a min count sketch where we can use the constant
   memory during collection phase, but this problem we only interested in one owner, i think hashmap can still work 

4. we can also parallel the last step with multiple go-routine and merge at the end, to accelerating the process (didn't implemente)
