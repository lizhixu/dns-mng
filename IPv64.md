**API URL**
https://ipv64.net/api
https://ipv64.net/api.php

**API Authentication**
`**Bearer Token:** Your Account API Key (Option 1)**Basic Auth Base64:** {xyz}:{apikey} - Your Account API Key (Option 2)**apikey|token:**[GET|POST] - Your Account API Key (Option 3)`

**API Limits**
`**API Account Limit:** API limit depends on the account class. (Default 64 / 24h)**API Call Limit:** Maximum 5 API requests within 10 sec.`

------

#### API Calls in the overview

**[GET] Account Informations**
`**[GET] get_account_info**`Get all the information about the user account.

**[GET] Account Logging**
`**[GET] get_logs**`Get the last 100 log entries.

**[GET] Domain Informations**
`**[GET] get_domains**`Get back all information regarding domains and records.

```
{
    "subdomains": {
        "wx1.api64.de": {
            "updates": 0,
            "wildcard": 1,
            "domain_update_hash": "aaa",
            "ipv6prefix": "",
            "dualstack": "",
            "deactivated": 0,
            "records": [
                {
                    "record_id": 603103,
                    "content": "1.1.1.1",
                    "ttl": 60,
                    "type": "A",
                    "praefix": "",
                    "last_update": "2026-04-13 04:16:29",
                    "record_key": "zzz",
                    "deactivated": 0,
                    "failover_policy": "0"
                }
            ]
        }
    },
    "info": "success",
    "status": "200 OK",
    "add_domain": "get_domains"
}
```

**[POST] Create Domain** (form-data)
`**[POST] add_domain** => Domainname [String Format: domainname.ipv64.net]`Creates a new domain and automatically creates A o. AAAA record.

**[DELETE] Delete Domain** (x-www-form-urlencoded)
`**[DELETE] del_domain** => Domainname [String Format: domainname.ipv64.net]`Domain will be deleted immediately with all known DNS records.

**[POST] Add DNS Record** (form-data)
`**[POST] add_record** => Domainname [String Format: domainname.ipv64.net]**[POST] praefix** => Domainpraefix [String Format]**[POST] type** => A,AAAA,TXT,CNAME,MX,NS,SRV [String Format]**[POST] content** => Content for DNS Record. [String Format]`A new DNS record is created in the specified domain.

**[DELETE] Delete DNS Record** (form-data)
`**[DELETE] del_record** => Domainname [String Format: domainname.ipv64.net]**[DELETE] praefix** => Domainpraefix [String Format]**[DELETE] type** => A,AAAA,TXT,CNAME,MX,NS,SRV [String Format]**[DELETE] content** => Content for DNS Record. [String Format]`ODER
`**[DELETE] del_record** => DNS Record ID [Integer Format]`The DNS record is immediately deleted from the domain.

------

**API-Response (Response)**

```
**HTTP-Header-Codes:** 200 OK, 201 Created, 400 Bad Request, 401 Unauthorized, 403 Forbidden, 429 Too Many Requests
```



```
**JSON-Payload:****Response:** Response to your API call**Status:** success OR http header code**API-Call:** Which call was called.
```



------

#### API Examples

**[GET] API Get Account Info:**
`curl -X GET https://ipv64.net/api.php?get_account_info -H "Authorization: Bearer 123456787654321234567876543"`



**[POST] API Create Domain:**
`curl -X POST https://ipv64.net/api.php -H "Authorization: Bearer 123456787654321234567876543" -d "add_domain=test1234.any64.de"`



**[DELETE] API Create Domain:**
`curl -X DELETE https://ipv64.net/api.php -H "Authorization: Bearer 123456787654321234567876543" -d "del_domain=test1234.any64.de"`