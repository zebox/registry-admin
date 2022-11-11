### Registry Admin
This  project allow manage repositories, images and users access  for self-hosted private docker registry with web UI. 
Main idea for implement this project to make high-level API for management user access to private registry 
and restrict their action (push/pull) for specific repositories (only with `token` auth) based 
on [official](https://docs.docker.com/registry/) private docker registry [image](https://hub.docker.com/_/registry). 
But someone need simple management to registry without split access to repositories. 
Registry Admin allow use either `password` or `token` authentication scheme for access management, 
depending on your task.  This application can be deployed with existed private registry for add access management 
UI tools to it.

Web user interface build with [React-Admin](https://marmelab.com/react-admin) framework and [MUI](https://mui.com/) components.

### Features
* Management users and access to registry
* Restrict access to repository by user action (`pull`/`push`, only for `token` auth scheme)


