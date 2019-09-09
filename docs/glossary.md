# Proctor Glossary

### Execution context
Execution context or context is a record of procs execution request by user to execute.
Context contain data such as:
  * procs name
  * context name
  * user email
  * procs tag
  * args given
  * procs output
  * procs status

### Procs
Procs is a job to execute using Proctor bundled as docker image and have set of metadata and secrets.

### Procs metadata
Procs metadata or mostly addressed as metadata contain metadata of procs such as:
  * name: Name of the procs
  * description: Description of the procs
  * author: Who create this procs
  * contributors: People that contribute to the procs
  * organization: Which org own this procs
  * env_vars:
    * secrets: Secret value that required by procs to run
    * args: Arguments that can be passed to procs

### Secret
Secret variable required to run procs such as credentials for cloud platform.
Proctor user shouldn't know the value of secret.
