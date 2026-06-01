---
name: fireblazer
description: Enumerates Google Cloud Platform (GCP) services accessible by an API key. Use when you need to determine the scope of a GCP API key (usually starting with 'AIza').
command: /fireblazer
---

# Fireblazer: AI Agent Tool Instructions

**Fireblazer** is an inspection utility for GCP API keys. It determines the scope of an API key by rapidly testing it against hundreds of Google API discovery endpoints. 

## When to Use This Tool
Use Fireblazer when the user provides a Google Cloud Platform API key (typically beginning with `AIzaSy...`) and asks to:
*   Identify the project associated with the key.
*   Enumerate the services or APIs accessible by the key.
*   Determine the scope of the key.
*   "Blaze" or scan the key.

## What Fireblazer Is and Is NOT
*   **IT IS an Enumeration Tool:** It uses the fact that Google's discovery endpoints verify project access.
*   **IT IS NOT an Exploit:** Discovering that a key has access to `datastore.googleapis.com` does not mean the database is vulnerable or exposed. It simply means the project utilizes that API, and the key is authorized to interface with it to some degree. The API itself may be further protected by IAM or OAuth gating. True vulnerabilities depend entirely on customer misconfigurations.
*   **Stealth:** This tool is **NOT stealthy**. It is just unlikely that it rings any alarms in the service, as it just causes a 403 response per enabled service for the google.discovery.Discovery method. This is not a concern for most API keys, and if you intend on any sort of enumeration in general, you'll do something similar anyways. As far as dynamic recon goes for GCP, this is fine. This tool is safe to use on any given API key.
*   **Blaze Mode (`-blaze`):** This flag performs additional recon (like fetching the OAuth screen brand identity or inferring service accounts). It is even noisier. Use it if the user requests deep recon or explicitly mentions blazing.

## Installation
If the `fireblazer` command is not available in your environment, install it via Go:
```bash
go install github.com/bedros-p/fireblazer@latest
```
*Note: Ensure `~/go/bin` is in your `$PATH`, or invoke it directly via `~/go/bin/fireblazer`.*

## Execution Rules (CRITICAL FOR AI AGENTS)

1.  **Format for Agents:** You **MUST** use `-outputFormat=text` (or `json`/`yaml`). The default `interactive` mode uses ANSI spinners which will corrupt your ability to read the output in the terminal. You will have many lines that are just progress updates. ~430 pointless lines per key.
2.  **Argument Ordering:** All configuration flags **MUST precede** the API key(s). 
3.  **Batch Processing:** Fireblazer supports scanning multiple keys simultaneously. Provide them as space-separated positional arguments at the end of the command.
4.  **Referrer Restrictions:** If the key is restricted to a specific origin or app, prefix the key using this syntax:
    *   Web: `xref:AIza...:example.com`
    *   iOS: `xios:AIza...:com.example.app`
    *   Android: `xandroid:AIza...:com.example.app:CERT_FINGERPRINT`

### Example Invocations

**Standard Text Scan:**
```bash
fireblazer -outputFormat=text AIzaSy...
```

**Aggressive Recon (Blaze Mode) formatted as JSON:**
```bash
fireblazer -blaze -outputFormat=json AIzaSy...
```

**Batch Scanning with Referrers:**
```bash
fireblazer -outputFormat=text AIzaSy_Key1 xref:AIzaSy_Key2:example.com xios:AIzaSy_Key3:com.app
```

## CLI Usage / Help Output
```text
Usage of fireblazer:
  -apiKey string
    	API key to scan. Can also be your first positional arg.
  -blaze
    	Enable additional aggressive recon checks (e.g., Brand Identity)
  -dangerouslySkipVerification
    	Skip API key verification (Note: This stops Blaze mode from working since it requires a Project ID)
  -findSlowService
    	[DEBUG] Find which service took the longest to test + elapsed time. Use to file an issue for program hangs.
  -outputDetails string
    	Comma delimited list of what to include in the details (description|title|name). (default "name")
  -outputFormat string
    	Output format (interactive|text|json|yaml) (default "interactive")
  -referrer string
    	Set the referrer (Referer:) header for when an API key is restricted to a website.
  -targetApi string
    	A single API discovery endpoint to test against. Bypasses the full scan.
  -workerCount int
    	Set the amount of worker threads to spawn for executing the requests (default 170)
```