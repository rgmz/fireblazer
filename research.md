## Notes

### Firebase iamPerms enum
`curl 'https://mobilesdk-pa.clients6.google.com/v1/projects/project-name-appspot-id:testIamPermissions?alt=json&key=KEY'` \
Get auth header from Firebase Console ? Anyways, there are better iam, this just stuck out because the request body had all the cool iam stuff

Post data: 
```json
{"permissions":["firebase.projects.own","firebase.projects.get","firebase.projects.update","firebase.projects.delete","resourcemanager.projects.update","resourcemanager.projects.delete","resourcemanager.projects.get","resourcemanager.projects.getIamPolicy","resourcemanager.projects.setIamPolicy","apikeys.keys.create","apikeys.keys.get","apikeys.keys.list","apikeys.keys.update","serviceusage.services.enable","serviceusage.services.get","firebase.clients.create","firebase.clients.delete","firebase.clients.get","firebase.clients.list","firebase.clients.undelete","firebase.clients.update","clientauthconfig.brands.create","clientauthconfig.brands.get","clientauthconfig.brands.update","clientauthconfig.clients.create","clientauthconfig.clients.delete","clientauthconfig.clients.get","clientauthconfig.clients.list","clientauthconfig.clients.update","oauthconfig.verification.get","firebase.billingPlans.get","firebase.billingPlans.update","resourcemanager.projects.createBillingAssignment","resourcemanager.projects.deleteBillingAssignment","firebaseapphosting.backends.create","firebaseapphosting.backends.delete","firebaseapphosting.backends.get","firebaseapphosting.backends.invoke","firebaseapphosting.backends.list","firebaseapphosting.backends.update","firebaseapphosting.builds.create","firebaseapphosting.builds.delete","firebaseapphosting.builds.get","firebaseapphosting.builds.list","firebaseapphosting.builds.update","firebaseapphosting.domains.create","firebaseapphosting.domains.delete","firebaseapphosting.domains.get","firebaseapphosting.domains.list","firebaseapphosting.domains.update","firebaseapphosting.locations.get","firebaseapphosting.locations.list","firebaseapphosting.operations.get","firebaseapphosting.operations.list","firebaseapphosting.operations.delete","firebaseapphosting.operations.cancel","firebaseapphosting.rollouts.create","firebaseapphosting.rollouts.delete","firebaseapphosting.rollouts.get","firebaseapphosting.rollouts.list","firebaseapphosting.rollouts.update","firebaseapphosting.traffic.get","firebaseapphosting.traffic.list","firebaseapphosting.traffic.update","firebaseauth.configs.create","firebaseauth.configs.get","firebaseauth.configs.getHashConfig","firebaseauth.configs.getSecret","firebaseauth.configs.update","firebaseauth.users.create","firebaseauth.users.get","firebaseauth.users.update","firebaseauth.users.delete","firebaseauth.users.sendEmail","firebasedatabase.instances.create","firebasedatabase.instances.get","firebasedatabase.instances.list","firebasedatabase.instances.update","firebasedatabase.instances.delete","datastore.databases.create","datastore.databases.delete","datastore.databases.getMetadata","datastore.databases.get","datastore.databases.list","datastore.databases.update","datastore.entities.get","datastore.entities.list","datastore.entities.create","datastore.entities.delete","datastore.entities.update","datastore.indexes.get","datastore.indexes.list","datastore.indexes.create","datastore.indexes.delete","datastore.indexes.update","datastore.backupSchedules.create","datastore.backupSchedules.delete","datastore.backupSchedules.list","datastore.backupSchedules.get","datastore.backupSchedules.update"]}
```


If identitytoolkit / identitytoolkit.googleapis.com is supported - https://www.googleapis.com/identitytoolkit/v3/relyingparty/getProjectConfig?key=AIzaKEY
```json
{
  "projectId": "101010101010101",
  "authorizedDomains": [
    "example-out-prod.firebaseapp.com",
    "example-out-prod.web.app",
    "exampleout.com",
    "pre-prod.exampleout.com",
    "output-example.redacted.com",
    "output-example-redacted.redacted.com",
    "mock.example.com",
    "localhost"
  ]
}
```

https://identitytoolkit.googleapis.com/v1/projects?key=
^ alias ? says it's legacy in the docs.

### Identitytoolkit Response Schema
https://identitytoolkit.googleapis.com/$discovery/rest
```json
 "IdentitytoolkitRelyingpartyGetProjectConfigResponse": {
  "id": "IdentitytoolkitRelyingpartyGetProjectConfigResponse",
  "description": "Response of getting the project configuration.",
  "type": "object",
  "properties": {
    "useEmailSending": {
      "description": "Whether to use email sending provided by Firebear.",
      "type": "boolean"
    },
    "changeEmailTemplate": {
      "$ref": "EmailTemplate",
      "description": "Change email template."
    },
    "projectId": {
      "type": "string",
      "description": "Project ID of the relying party."
    },
    "resetPasswordTemplate": {
      "$ref": "EmailTemplate",
      "description": "Reset password email template."
    },
    "authorizedDomains": {
      "type": "array",
      "description": "Authorized domains.",
      "items": {
        "type": "string"
      }
    },
    "legacyResetPasswordTemplate": {
      "$ref": "EmailTemplate",
      "description": "Legacy reset password email template."
    },
    "dynamicLinksDomain": {
      "type": "string"
    },
    "verifyEmailTemplate": {
      "$ref": "EmailTemplate",
      "description": "Verify email template."
    },
    "enableAnonymousUser": {
      "description": "Whether anonymous user is enabled.",
      "type": "boolean"
    },
    "apiKey": {
      "description": "Browser API key, needed when making http request to Apiary.",
      "type": "string"
    },
    "idpConfig": {
      "items": {
        "$ref": "IdpConfig"
      },
      "type": "array",
      "description": "OAuth2 provider configuration."
    },
    "allowPasswordUser": {
      "description": "Whether to allow password user sign in or sign up.",
      "type": "boolean"
    }
  }
}
```

### Uses
Could be used for extracting API key - project link regardless. If the key has access to the endpoint, we get the project id. If it doesn't... we get the project id anyways! I was looking for an endpoint that is always guaranteed to return the project id.

Other approaches rely on just error codes, could fail if the api key DOES have access to the method. Even Google Private APIs / -pa.googleapis would fail to give the project ID if the site you're testing is owned by google. Though... Most of those have referrer restrictions so it might still give it. Either way, in this case, both the error code and response include it, so it's perfect. No edge cases as far as I know ((?)).

### Responses & err codes
```json
{
  "error": {
    "code": 403,
    "message": "Requests to this API identitytoolkit method google.cloud.identitytoolkit.v1.ProjectConfigService.GetProjectConfig are blocked.",
    "errors": [
      {
        "message": "Requests to this API identitytoolkit method google.cloud.identitytoolkit.v1.ProjectConfigService.GetProjectConfig are blocked.",
        "domain": "global",
        "reason": "forbidden"
      }
    ],
    "status": "PERMISSION_DENIED",
    "details": [
      {
        "@type": "type.googleapis.com/google.rpc.ErrorInfo",
        "reason": "API_KEY_SERVICE_BLOCKED",
        "domain": "googleapis.com",
        "metadata": {
          "service": "identitytoolkit.googleapis.com",
          "methodName": "google.cloud.identitytoolkit.v1.ProjectConfigService.GetProjectConfig",
          "consumer": "projects/133333337",
          "apiName": "identitytoolkit"
        }
      },
      {
        "@type": "type.googleapis.com/google.rpc.LocalizedMessage",
        "locale": "en-US",
        "message": "Requests to this API identitytoolkit method google.cloud.identitytoolkit.v1.ProjectConfigService.GetProjectConfig are blocked."
      }
    ]
  }
}
```

## Thoughts...
This also means the key has  "https://www.googleapis.com/auth/cloud-platform" as a scope (most keys do though)

IAM Permission endpoint seems to be all over the place though. Look for RPC endpoint that can be standard.

On Google Cloud Storage, it's https://www.googleapis.com/storage/v1/b/bucket-name-here/iam/testPermissions?permissions=... but on ResourceManager it's on /v1/projects/PROJECT_ID_ALPHANUMERIC:testIamPermissions

https://gcp.permissions.cloud/ looks like a scraped IAM list. I dont think this is needed, as i have plenty of endpoints that list it. Might be more comprehensive, but this is pretty convenient for a starter.

Seems to be powered by https://github.com/iann0036/iam-dataset

```
map.json
A map of IAM permissions required for each method. [WORK IN PROGRESS]
```

I need a consistent IAM testperm endpoint, above all else. Seems like a good project though, might include a link to it in the README as a sorta "also check out...". Can't integrate it currently as it's not what I need. \
It seems to operate off the autogenerated Google Cloud docs, but those can be unreliable / outdated.

Service management, usagemetrics and discovery docs + proper firstparty, non-scraped iam sources will work better here. And will be more up to date.

GAPIs supports listing buckets by project https://storage.googleapis.com/storage/v1/b?project=901333333337&key=AIzaSy... \
^ can test if API key has access to list buckets. Huge if any API key gets this.

There is no IAM bleed between services. Attempt at accessing logging.buckets.get through GCS:
```json
{
  "error": {
    "code": 400,
    "message": "logging.buckets.get is not a valid Google Cloud Storage permission.",
    "errors": [
      {
        "message": "logging.buckets.get is not a valid Google Cloud Storage permission.",
        "domain": "global",
        "reason": "invalid"
      }
    ]
  }
}
```

Might be worthwile to look into all GCP services that support a project query param / numeric project ids. Could provide additional info IAM does not.

The Github thing includes a pretty flattened api listing with perms in methods_ext. Could use it to find the ?project params, but I could also do that through my own existing archive of Discovery documents. Most services are covered in mine, even internal ones, along with ~3 revisions per discovery doc.

IAM test perms seems to be a common RPC endpoint, available like so - google.storage.control.v2.StorageControl/TestIamPermissions. \
I used to have an endpoint that could call any RPC provided the auth was right. No idea where it is now.

The IAM RPC shows up in googleapis for servicemanagement

----

I'll be prioritizing alternative output formats once the major features are all implemented. That way I don't stall features while I make output formats for each one.

https://www.googleapis.com/storage/v1/b/cloud-samples-data/iam/testPermissions?permissions=storage.buckets.delete&permissions=storage.buckets.get&key=KEY

The cloud-samples-data bucket is widely used in Google documentation.\
 If one day that bucket retires and someone else claims it, they'd have bigger problems :P
https://www.googleapis.com/storage/v1/b/cloud-samples-data/iam/testPermissions?permissions=storage.buckets.delete&permissions=storage.buckets.get&key=KEY


The bucket name itself is not important. This API actually has a better use! It tells you if the API key is tied to an active billing account :)

```json
{
  "error": {
    "code": 403,
    "message": "The billing account for the owning project is disabled in state absent",
    "errors": [
      {
        "message": "The billing account for the owning project is disabled in state absent",
        "domain": "global",
        "reason": "accountDisabled",
        "locationType": "header",
        "location": "Authorization"
      }
    ]
  }
}```

Interestingly... It doesn't check for ACL for the specified bucket for the storage.buckets stuff, it checks the API key project instead. I'm not so sure for the non bucket* ones.

Some keys gave me:
```json
{
  "kind": "storage#testIamPermissionsResponse"
}
```

While some other keys gave me:
```json
{
  "kind": "storage#testIamPermissionsResponse",
  "permissions": [
    "storage.buckets.get"
  ]
}
```

Could be very useful. Will add with checks for all the diff permissions.

```json
  "TestIamPermissionsResponse": {
   "id": "TestIamPermissionsResponse",
   "type": "object",
   "description": "A storage.(buckets|objects|managedFolders).testIamPermissions response.",
   "properties": {
    "kind": {
     "type": "string",
     "description": "The kind of item this is.",
     "default": "storage#testIamPermissionsResponse"
    },
    "permissions": {
     "type": "array",
     "description": "The permissions held by the caller. Permissions are always of the format storage.resource.capability, where resource is one of buckets, objects, or managedFolders. The supported permissions are as follows:  \n- storage.buckets.delete - Delete bucket.  \n- storage.buckets.get - Read bucket metadata.  \n- storage.buckets.getIamPolicy - Read bucket IAM policy.  \n- storage.buckets.create - Create bucket.  \n- storage.buckets.list - List buckets.  \n- storage.buckets.setIamPolicy - Update bucket IAM policy.  \n- storage.buckets.update - Update bucket metadata.  \n- storage.objects.delete - Delete object.  \n- storage.objects.get - Read object data and metadata.  \n- storage.objects.getIamPolicy - Read object IAM policy.  \n- storage.objects.create - Create object.  \n- storage.objects.list - List objects.  \n- storage.objects.setIamPolicy - Update object IAM policy.  \n- storage.objects.update - Update object metadata. \n- storage.managedFolders.delete - Delete managed folder.  \n- storage.managedFolders.get - Read managed folder metadata.  \n- storage.managedFolders.getIamPolicy - Read managed folder IAM policy.  \n- storage.managedFolders.create - Create managed folder.  \n- storage.managedFolders.list - List managed folders.  \n- storage.managedFolders.setIamPolicy - Update managed folder IAM policy.",
     "items": {
      "type": "string"
     }
    }
   }
  }
```

### Perms list for storage googleapis
storage.buckets.delete
storage.buckets.get
storage.buckets.getIamPolicy
storage.buckets.create
storage.buckets.list
storage.buckets.setIamPolicy
storage.buckets.update
storage.objects.delete
storage.objects.get
storage.objects.getIamPolicy
storage.objects.create
storage.objects.list
storage.objects.setIamPolicy
storage.objects.update
storage.managedFolders.delete
storage.managedFolders.get
storage.managedFolders.getIamPolicy
storage.managedFolders.create
storage.managedFolders.list
storage.managedFolders.setIamPolicy

### Perm URL

> Note: Providing storage.buckets.list or storage.buckets.create returns an error, as these permissions apply to projects instead of buckets\
> *from https://docs.cloud.google.com/storage/docs/json_api/v1/buckets/testIamPermissions*

We'll test for .list and .create using the ?project= parameter & running their respective API stuff.

https://docs.cloud.google.com/storage/docs/json_api/v1/buckets/insert

https://www.googleapis.com/storage/v1/b/cloud-samples-data/iam/testPermissions?permissions=storage.buckets.delete&permissions=storage.buckets.get&permissions=storage.buckets.getIamPolicy&permissions=storage.buckets.create&permissions=storage.buckets.list&permissions=storage.buckets.setIamPolicy&permissions=storage.buckets.update&permissions=storage.objects.delete&permissions=storage.objects.get&permissions=storage.objects.getIamPolicy&permissions=storage.objects.create&permissions=storage.objects.list&permissions=storage.objects.setIamPolicy&permissions=storage.objects.update&permissions=storage.managedFolders.delete&permissions=storage.managedFolders.get&permissions=storage.managedFolders.getIamPolicy&permissions=storage.managedFolders.create&permissions=storage.managedFolders.list&permissions=storage.managedFolders.setIamPolicy&key=

Having any sort of perm could very well be indicative of GCS being accessible through it. The above link has all the perms, this one is restricted to the excluded one:
https://www.googleapis.com/storage/v1/b/cloud-samples-data/iam/testPermissions?permissions=storage.buckets.delete&permissions=storage.buckets.get&permissions=storage.buckets.getIamPolicy&permissions=storage.buckets.setIamPolicy&permissions=storage.buckets.update&permissions=storage.objects.delete&permissions=storage.objects.get&permissions=storage.objects.getIamPolicy&permissions=storage.objects.create&permissions=storage.objects.list&permissions=storage.objects.setIamPolicy&permissions=storage.objects.update&permissions=storage.managedFolders.delete&permissions=storage.managedFolders.get&permissions=storage.managedFolders.getIamPolicy&permissions=storage.managedFolders.create&permissions=storage.managedFolders.list&permissions=storage.managedFolders.setIamPolicy&key=

Interestingly, I get this error for the cloud-samples-data bucket
```json
{
  "error": {
    "code": 400,
    "message": "Cannot test storage.objects.getIamPolicy or storage.objects.setIamPolicy on buckets with uniform bucket-level access enabled",
    "errors": [
      {
        "message": "Cannot test storage.objects.getIamPolicy or storage.objects.setIamPolicy on buckets with uniform bucket-level access enabled",
        "domain": "global",
        "reason": "invalid"
      }
    ]
  }
}
```

Uniform bucket-level access disables ACLs (Access Control Lists) - "access to Cloud Storage resources then is granted exclusively through IAM"
Neat way to check for uniform bucket-level access on a specified bucket for post-scan scripts.

There's also another public bucket thatr DOESNT use uniform bucket level access. 
`https://www.googleapis.com/storage/v1/b/gcp-public-data-landsat/iam/testPermissions?permissions=storage.buckets.delete&permissions=storage.buckets.get&permissions=storage.buckets.getIamPolicy&permissions=storage.buckets.setIamPolicy&permissions=storage.buckets.update&permissions=storage.objects.delete&permissions=storage.objects.get&permissions=storage.objects.getIamPolicy&permissions=storage.objects.create&permissions=storage.objects.list&permissions=storage.objects.setIamPolicy&permissions=storage.objects.update&permissions=storage.managedFolders.delete&permissions=storage.managedFolders.get&permissions=storage.managedFolders.getIamPolicy&permissions=storage.managedFolders.create&permissions=storage.managedFolders.list&permissions=storage.managedFolders.setIamPolicy&key=`

I do think anything past buckets.* is pointless on a random bucket not belonging to the key. Still, the IAM testPermissions evidently can be used as a way to get some sort of additional data.

Referring back to this - https://storage.googleapis.com/storage/v1/b?project=PROJECT-ID-NUMBERS&key=AIza

When I checked it against my own project id and key, for a project that does not have GCS:
```json
{
  "error": {
    "code": 401,
    "message": "Requests to this API storage.googleapis.com method google.storage.buckets.list are blocked.",
    "errors": [
      {
        "message": "Requests to this API storage.googleapis.com method google.storage.buckets.list are blocked.",
        "domain": "global",
        "reason": "required",
        "locationType": "header",
        "location": "Authorization"
      }
    ]
  }
}
```

I don't have a billing account on GCP. If I did, my research would be greatly accelerated. Such is life. Maybe if my recent bugs pay out I might invest a good bit into my cybersec research. But if I get even more tea it would be just as good at helping me out in cybersec... too many options...

-----
