This project consists of uploading, download, list and exclude files to a shared folder on Google Drive.

To use the project, you need to register an API project on Google Cloud https://console.cloud.google.com

Menu - APIs and services enabled:
- Activate the options listed below:
  - Google People API
  - Google Drive API

Menu - Credentials:
Create your credentials file in the "Service Accounts" option. Once created, download and save the .JSON file in your application folder.
In the main.go file, change the ServiceAccount constant to the path where your file is located, as in the example:
  ServiceAccount = "C:\\Projects\\GO\\UploadFile\\Credential.json"

Once this part of the configuration is done, in your Google Drive, create a folder and share it with the email address generated when you created the authentication in "Service Accounts".

Finally, when sharing the folder, copy the name of the folder you created and change the property "Parents: []string{"yourfolderpath"}" in the main.go file
Example of what the URL "https://drive.google.com/drive/folders/yourfolderpath" looks like

After completing these steps, you can run the main.go file with the command "go run main.go", which will open a dialog for you to select the desired file and perform the update.
