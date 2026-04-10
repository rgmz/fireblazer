# Fireblazer

Extract all services used by a Google Cloud Platform project with a regular API key like "AIza...".\
Good for expanding your scope from a mere Firebase key to every service that may be unprotected.

This program does not take rely on any vulnerabilities. It is an INSPECTION UTILITY, *not an exploit*. Pentesters and bug hunters are the intended users. More in the "NOT AN EXPLOIT" section.

## Installation
```bash
go install github.com/bedros-p/fireblazer@latest
```

### From source

```
git clone https://github.com/bedros-p/fireblazer
go mod download
go build .
./fireblazer -h
```

`go build .` creates a binary `fireblazer`, what happens after that is up to you :)

## Usage
Example usage & output\
`fireblazer AIzaSyC334f24LundukeS8uSkjWoke18`

Output:
```log
2026/04/02 20:33:16 Successfully loaded 501 discovery endpoints from hardcoded list.
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

The program also checks the validity of the API key. If you're confident it's valid / want to save .2 seconds on the ~5 second scan, use --dangerouslySkipVerification. It's not really for saving time, but in case the primary verification method is broken.

Enjoy the API key escalation!

## Roadmap / Plans
### Major Features
- Support multiple output formats (YAML, JSON, Plain text & fancy cli outputs \[spinners\]) (Partial implementation)
- Show which services require OAuth & which require Service Accounts to prevent the pentester from wasting time
- ^ Related, IAM testing on all endpoints through /iam/testPermissions would result in an even greater reduction in time necessary.
- Suggested actions & quick execs (firebase bucket perm testing)
- Include flag to check for autopush, staging, preprod and -pa variations of the APIs. Only useful for testing Google owned keys, so it's kind of a personal want.

### Patches
- Add special detection methods for the (filtered out) false positives (refer to false positives from main.go) - priority would be the GCS API.
- Timestamps to be disabled by default
- Sort keys by service count

#### Bugs 
- If identitytoolkit is enabled but not configured, fireblazer will break on validity checking - "400, configuration not found". Critical issue !
- The remaining counter tends to be unstable as new keys are added to the scan (looks very jittery). Simple fix.

## NOT AN EXPLOIT

> This program is NOT an exploit in any way.

This whole program relies on the fact that a Discovery endpoint is almost guaranteed to be there on every PUBLIC Google service endpoint, and that it still checks if the project associated with the key uses the intended service for it. This is NOT a design flaw, nor is it a problem.

Google cloud will warn you - restrict your keys. IT even has a big yellow warning sign telling you to restrict it when you make it.

This can't be avoided by any reasonable measure - if Discovery URLs didn't check for key validity, one could easily test each service with one of the actual endpoints. We do, after all, have a list of all the actual endpoints in a service. It would still show whether or not the API key is used in that project, regardless of an invalid payload to the endpoint. Checking if the endpoint payload is valid before checking for the API key is not possible, as most services tie the project data into different responses, and the project ID is inferred from the API key. It would require each service to have a rewrite of checks.\
Or, they can return zero errors of use and make it silently fail, making life hell for the people that actually want to develop with GCP, in an attempt to safeguard information that doesn't pose much of a security risk. The real security risk is entirely dependent on how securely the project is set up. Just additional safeguarding that ruins DX for something that wouldn't be the root problem anyways.

## Notes 

Uses HTTP3 (QUIC) for less cancelled / retransmitted requests, it's faster. On inferior versions, this would retransmit lots of packets unnecessarily. You can test out the error rate by switching out http3.Transport to a regular http.Transport in `client.go`

The code isn't the best quality. It's been a while since I've done Go, and I didn't want to use AI for this. Not yet anyways. Trying to regain brain function after months of TS only development. Though I suppose Go isn't too far away in terms of brain usage. Improvements are VERY welcome, even the nits. Though, if you're gonna raise an issue for nits, please combine all nits into one issue, just optimization and better practices. 

Only reason I'd want to merge a PR for nits is if it came with an item on the roadmap too. Otherwise, post it as an issue, and I'll get to it :)

This took too long to make for such little code.