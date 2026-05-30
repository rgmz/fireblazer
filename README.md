# Fireblazer

Extract all services used by a Google Cloud Platform project with a regular API key like "AIza...".\
Good for expanding your scope from a mere Firebase key to every service that may be unprotected. Supports enumeration against 440 Google Cloud services.

This program does not take rely on any vulnerabilities\*. It is an INSPECTION UTILITY, *not an exploit*. Pentesters and bug hunters are the intended users. More in the "NOT AN EXPLOIT" section.

> More specifically, this doesn't rely on anything that Google considers a vulnerability. The service listing is not actionable. If it finds an attack vector, that is entirely on the customer project.

## Installation
```bash
go install github.com/bedros-p/fireblazer@latest
```

## Usage
Example usage & output\
`fireblazer AIzaSyC334f24LundukeS8uSkjWoke18`

Output:
```log
2026/04/02 20:33:16 Successfully loaded 440 discovery endpoints from hardcoded list.
✓ [AIzaSyC334f24LundukeS8uSkjWoke18] Valid API key, proceeding.
✓ Scan complete!
2026/04/02 20:33:20 APIs available to this API key with project ID 30507080705752:
2026/04/02 20:33:20  - cloudasset.googleapis.com
2026/04/02 20:33:20  - datacatalog.googleapis.com
2026/04/02 20:33:20  - containeranalysis.googleapis.com
2026/04/02 20:33:20  - datastore.googleapis.com
2026/04/02 20:33:20  - dataform.googleapis.com
2026/04/02 20:33:20  - container.googleapis.com
```

Batch usage is supported through positional arguments. If you are testing multiple keys, this is highly recommended. I've had a speed up of roughly ~18x when verifying 100 keys.
It starts to struggle around the 300 mark, so be careful with your quantity.

`fireblazer AIzaSyC334f24LundukeS8uSkjWoke18 AIzaSyC334fSkafGr4h5ke485Sk25okt12` - chain it as much as you want. I don't recommend chaining more than 100 at a time - your mileage may vary.

If you have keys that are origin restricted (like to an Android app, iOS app, or website), you can pass it through like so:

- `xref:KEY:example.com` - Sets the `Referer` header.
- `xios:KEY:com.example.app` - Sets the `X-Ios-Bundle-Identifier` header.
- `xandroid:KEY:com.example.app:CERT` - Sets the `X-Android-Package` and `X-Android-Cert` headers. Make sure to strip the colons from the cert!

Example: `fireblazer xios:AIzaSyC334f24...:com.google.gemini xref:AIzaSy...:gemini.google.com`

I went with this format as it works in regular and batch mode. Also, if you're unsure about whether or not it's restricted, if it's information you have, no harm in including it.

You can also specify a single API endpoint to send a GET request against, with the --targetApi command
`fireblazer --targetApi=www.googleapis.com/discovery/v1/apis/drive/v3/rest AIzaSyD...`

Useful for when you only care about checking a batch of keys with an http3 stream, it'll be faster than making another script. Less errors, more reliable results, faster overall :)

The program also checks the validity of the API key. If you're confident it's valid / want to save .2 seconds on the ~5 second scan, use --dangerouslySkipVerification. It's not really for saving time, but in case the primary verification method is broken. Please file an issue or mention me on Twitter if that's the case.

You can also change the output format using the `--outputFormat` flag. The available options are `interactive`, `text`, `json`, and `yaml`. This is especially useful for integrating Fireblazer into automated pipelines.
Example: `fireblazer --outputFormat=json AIzaSy...`

Enjoy the API key escalation!

## Build

```
git clone https://github.com/bedros-p/fireblazer
go mod download
go build .
./fireblazer -h
```

`go build .` creates a binary `fireblazer`, what happens after that is up to you :)

Improvements are VERY welcome, even the nits. Though, if you're gonna raise an issue for nits, please combine all nits into one issue, just optimization and better practices. Only reason I'd want to merge a PR for nits is if it came with an item on the roadmap too. Otherwise, post it as an issue, and I'll get to it :)

The base is entirely handwritten by me. I may use AI for code cleanup, but otherwise it's best for stability if the core is human maintained (for now). Any AI involvement should consist of inline completions for the sake of efficiency in place of IDE snippets. If you intend on submitting AI-only PRs, avoid it. This tool uses lots of aspects of Google that took days of research.

## Roadmap / Plans
### Major Features
- Show which services require OAuth & which require Service Accounts to prevent the pentester from wasting time
- ^ Related, IAM testing on all endpoints through /iam/testPermissions would result in an even greater reduction in time necessary.
- Suggested actions & quick execs (firebase bucket perm testing)
- Include flag to check for autopush, staging, preprod and -pa variations of the APIs. Only useful for testing Google owned keys, so it's kind of a personal want.
- Add Brand Identity retrieval thru clientauthconfig.googleapis.com w the key from https://console.cloud.google.com (upon running into this and looking into it further, turns out its been reported as a vulnerability and marked as wontfix - https://feed.bugs.xdavidhu.me/bugs/0009 )
- Minor, but a tuning mode to identify the ideal worker count for future runs on the machine it's on.

### Patches
- Add special detection methods for the (filtered out) false positives (refer to false positives from main.go) - priority would be the GCS API.
- Sort keys by service count
- Investigate multipart batch calls. The performance gain would be minimal, but it's interesting. I scrapped that a while back as http3 allows sending it all as "one request" (stream) anyways, but it would be interesting if we can minimize any HTTP overhead, no matter how slim HTTP3 is. Just an experiment. Some thoughts in Notes.

#### Bugs 
- If identitytoolkit is enabled but not configured, fireblazer will break on validity checking - "400, configuration not found". Critical issue... but can't reproduce since? If anyone can repro pls file an issue!

## NOT AN EXPLOIT

> This program is NOT an exploit in any way.

This whole program relies on the fact that a Discovery endpoint is almost guaranteed to be there on every PUBLIC Google service endpoint, and that it still checks if the project associated with the key uses the intended service for it. This is NOT a design flaw, nor is it a problem.

Google cloud will warn you - restrict your keys. IT even has a big yellow warning sign telling you to restrict it when you make it.

This can't be avoided by any reasonable measure - if Discovery URLs didn't check for key validity, one could easily test each service with one of the actual endpoints. We do, after all, have a list of all the actual endpoints in a service. It would still show whether or not the API key is used in that project, regardless of an invalid payload to the endpoint. Checking if the endpoint payload is valid before checking for the API key is not possible, as most services tie the project data into different responses, and the project ID is inferred from the API key. It would require each service to have a rewrite of checks.\
Or, they can return zero errors of use and make it silently fail, making life hell for the people that actually want to develop with GCP, in an attempt to safeguard information that doesn't pose much of a security risk. The real security risk is entirely dependent on how securely the project is set up. Just additional safeguarding that ruins DX for something that wouldn't be the root problem anyways.

### Additional recon [WIP]

Entirely WIP, just here to solidify the roadmap so I'd finally do it today. If this is still not implemented by the time you're reading this, pester me on Twitter @ https://x.com/bedros_p

When enumerating a single key in interactive mode, it runs some additional, non-state-altering recon in its newest version, extending beyond mere API key + service pairing. This means that not all discovered services are particularly reachable by the API key, but exist merely for enumerating a projects infrastructure. These are listed in a separate segment. If it can be found through other means, Google VRP does not accept it as a concern. If it's not a risk on it's own, it's not a risk till a chain can be proven. In this case, all the chains are customer-specific, and I have not found anything that leads me to believe this relies on misconfigurations on Googles end. None of this is really a risk.

- [WIP, not impld] Brand lookup for OAuth screen (previously reported, marked as wontfix - https://feed.bugs.xdavidhu.me/bugs/0009 )
- [WIP, not impld] Check for service account existence based on well-known formats. For some specific services, enabling it causes the creation of a service account which we can check the existence of. The key cannot hit those services, it is only useful for recon on the project as a whole. This method is useful for target scoping, but poses absolutely zero threat to any customer. Thinking it's a problem is like saying "we can tell you use Spring Boot based on these error responses", but even less useful.

You can enable these checks for all scans if you like it with the -blaze flag.

## Notes 

Uses HTTP3 (QUIC) for less cancelled / retransmitted requests, it's faster. On inferior versions, this would retransmit lots of packets unnecessarily. You can test out the error rate by switching out http3.Transport to a regular http.Transport in `client.go`

This was originally made with the intent of enumerating the scope of Firebase projects. This took too long to make for such little code, but it's optimized and thoroughly tested :)

Back with HTTP2 when I was dealing with another bruteforce experiment, I would increase the ulimit and use the ephem port range when bruteforcing heavily. I'd run out, and I'd just add another network interface. My router holds a grudge against me to this day. Still, I do wonder what is the equivalent of an additional net interface is in this scenario though, would be cool if I could get it to run *even* faster. I haven't really checked for the QUIC stream limit Google has defined, it could be the case that quic-go is handling data restrictions under the hood as a retry when it would be better as another stream. Their auto-scaling logic works pretty well though... Later on I'll test pushing the defaults quic-go has against a testrun where apis.go is just duplicates of an enabled service and see if any get dropped when I push it to the max.

Originally I was gonna write my own HTTP3 client that crafts the raw packets. Decided against it for the sake of getting this one thing done and winning against my attention span, but feel free to put up a PR if you find it brings better speeds!

Enjoy using Fireblazer, and I hope it helps you with recon!