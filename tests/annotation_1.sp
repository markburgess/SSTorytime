
- test annotation

  Long string >>extractme with insert  // 'extractme', 'Long string extractme with insert'

  Other string >>"quoted insert"       // '"quoted insert"', 'Other string "quoted insert"'

  Doo bee >>`doh boh doh` bee doo      // 'doh boh doh', 'Doo bee doh boh doh bee doo'

  Test >>URL "https://domain1.com?q=test" // 'URL', 'Test URL "https://domain1.com?q=test"'
  
  one two >>three456 xyz               // 'three456', 'one two three456 xyz'

  alpha >>`bully bully boo`3456 to you  // 'bully bully boo', 'alpha bully bully boo3456 to you'

  nothing to do   

