# Passwordless Authenticator Proof-of-Concept

![Go Status](https://github.com/irby/passwordless-authenticator-poc/actions/workflows/go.yml/badge.svg)
![Angular Status](https://github.com/irby/passwordless-authenticator-poc/actions/workflows/angular.yml/badge.svg)

## Allow users to securely share their accounts in a passwordless environment


#### Contributors:
- Matthew H. Irby (mirby7@gatech.edu)

#### Related Projects
- Hanko ([repo](https://github.com/teamhanko/hanko)) - The authenticator backend of this project is branched off of this project.


<br/>


![Password Logo](https://securityintelligence.com/wp-content/uploads/2018/10/si-eight-character-password-feature.jpg)

<br/>


This project represents my capstone project for Fall 2022 at the Georgia Institute of Technology. All work here is for educational and research purposes, and it is not yet suited for production usage.

# Concept

After many years of design and gradual adoption, passwordless authentication is now being adopted by large companies such as Mirosoft, Apple, and Google. Passwordless authentication is designed to eliminate a large target for malicious users: passwords. 

As it is currently designed, passwordless authentication does not allow users to be able to share their accounts with others. I would like to introduce a novel concept of allowing users to securely share their accounts without risk of account takeover, and also allow IT auditing to identify which individual is accessing an account.

# Project Structure

The project contains the following structure:

- ### Authenticator Backend
  - This contains the code to run the authenticator API. The authenticator API will be consume by the authenticator front end, client front-end and client back-end.

- ### Authenticator Frontend
  - The contains the code for the authenticator front-end that will be run by two different users on differing sessions: the primary account holder and another user the account holder will share the account with.

- ### Client API
  - This contains the code to run the client app's backend. The client API will send the authentication token to the authenticator API to validate session tokens and send refresh tokens. The API will also handle logic of its own, such as serving its content to the client front-end.

- ### Client Frontend
  - This contains the code for the front-end that will be consumed by the primary account owner and the guest. It will have a login page and a page that requires a valid session token to access.

- ### Deploy
  - This contains the files necessary to spin up a Docker instance of the project directory. It will create an instance of the authenticator API, two instances of the authenticator front-end (for the primary account user and guest user), two instances of the client front-end (for the primary account user and guest user), one instance of the client API, a PostgreSQL database, and a mail server (mailslurper).

  - Use the following command to start the Docker container:
  
  ```(shell)
  docker-compose -f deploy/docker-compose/quickstart.yaml -p "hanko-quickstart" up --build
  ```

  - When running, the services can be reached at the following URLs:
    - Authenticator Backend - http://localhost:8000
    - Mailslurper - http://localhost:8080
    - Client Frontend - http://localhost:4200


# Accounts
The following accounts are scaffolded when the project is built out with Docker Compose:
- mirby7@gatech.edu
- gburdell27@gatech.edu
- buzz@gatech.edu
