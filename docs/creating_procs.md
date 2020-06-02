# Creating procs

Main purpose of proctor is running procs so you need to know how to create procs.
Here are the recommended steps to create procs.

#### 1. Define the task you want to automate.
Candidate task for a procs usually involving access to restricted resources or having complicated steps.

#### 2. Define the interface of the task.
A task takes input, performs operation and provide output.

#### 3. Research how to automate the task
Before start writing the script, please do research on how to do the automation by reading the documentation or stuff.

#### 4. Write the script
Because procs will be run by executor(which is computer) you need to write down the script.
Write a script such that it's runnable on your local machine. For this step you can hardcode on input, we'll extract it out later.

#### 5. Test the script
Don't skip this step, please test it on your local.

#### 6. Package script in a docker image
Proctor leverage container to easily run task with all it's dependency, that's why you need to package your script into a docker image.
Install every dependency on your dockerfile. Provide an `ENTRYPOINT` to run the script by default when the container is spun up from the image.

#### 7. Extract all hard-coded variables and secrets as configurable ENV vars
The image are meant to be reusable so user can use it according their use case, this is why every variables that define the behaviour of the automation should be extracted as args.

#### 8. Perform validations on ENV vars
Some ENV vars are mandatory in order to use the automation, adding validation before running script helps failing fast and provide better error messages to user.

#### 9. Test the image (Pass all variables as env)
Build the image then run it in local docker to make sure your image run as expected.
After you complete this step, your task can be automated using docker on any machine, post these steps you automation will evolve into a `proc`

#### 10. Create metadata
Create metadata file to describe information about your procs to user.
Metadata look like this;
```json
{
    "name": "echo-worker",
    "description": "This procs will echo your name",
    "image_name": "walbertusd/echo-worker",
    "env_vars": {
        "secrets": [
            {
                "name": "SECRET_NAME",
                "description": "My other secret name"
            }
        ],
        "args": [
            {
                "name": "NAME",
                "description": "Name to be echoed"
            }
        ]
    },
    "authorized_groups": [
        "my-group"
    ],
    "author": "Dembo",
    "contributors": "Dembo",
    "organization": "GoJek"
}
```

#### 11. Upload metadata
Send `POST` request to your proctor service on `<proctor-host>/metadata`, it receive array of metadata as json so your request body should look like this:
```json
[
    {
        "name": "echo-worker",
        "description": "This procs will echo your name",
        "image_name": "walbertusd/echo-worker",
        "env_vars": {
            "secrets": [
                {
                    "name": "SECRET_NAME",
                    "description": "My other secret name"
                }
            ],
            "args": [
                {
                    "name": "NAME",
                    "description": "Name to be echoed"
                }
            ]
        },
        "authorized_groups": [
            "my-group"
        ],
        "author": "Dembo",
        "contributors": "Dembo",
        "organization": "GoJek"
    }
]
```

#### 12. Store secret
User aren't supposed to know the secret value to run your jobs so make sure to keep it secret.
Send `POST` request to your proctor service on `<proctor-host>/secret`, your request body will look like this:
```json
{
	"job_name": "echo-worker",
	"secrets": {
		"SECRET_NAME": "Iron Man"
	}
}
```

#### 13. Test your proctor using CLI
Execute your procs using CLI, make sure it success and the resulting log is correct.

#### 14. Complete
Congratulations! You've just automated one repetitive task!
