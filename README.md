# Go API Server

This is a small backend for personal portfolio websites. I tried to dynamically manage features like Github syncing and contact form processing, while also being stable and secure all in a lightweight and fast package, i also added in goldmark's Markdown to HTML Converter library, as i update my about text quite frequently and i can do so without editing other vital files.

I created this, since i wanted a backend for my personal page, there's still a few To-Do's left for now: 

* **Rate Limiting -** I currently handle it by using NginX but i want to have it built-in eventually  
* **CI/CD -** Setting up CI/CD to automatically run build checks and tests  
* **Unit/Integration Tests -** Adding tests for the HTTP Handlers and Clients using standard and http packages

Feel free to use this under MIT License and if you have some suggestions on what i can do better, i would love some messages :)

That being said, lets go over some core features:

* **Caching -** Github Projects are cached directly by a dedicated go routine in memory every 20 minutes to avoid overloading the GitHub API and enable seamless usage by visitors of your page, while still being up to date, which also allows for a more zero latency approach for said visitors. The about.md file is also cached in RAM and updated every 60 minutes via a Go Routine.
* **Security -** I am avoiding Go's default http client to add explicit timeouts, so there's additional security in terms of request attacks and also the CORS Handler, aswell as Multiplexer is set to only accept the needed methods of the Github / Resend API respectively. Also i added graceful shutdowns to cleanly close active connections and finish ongoing requests before shutting down.
* **Separation of Concerns -** I tried to use the "Go Standard" way of organizing the project, aswell as adhere to Separation of Concerns Design principles to have stuff working as independantly as possible.
  

## Prerequisites for using this project
Make sure to read through the about.go and projects.go files comments if you want to change the frequency of updates or want to update the unsafe mode for the Markdown to HTML converter, which allows you to embed html tags, use your css classes and embed JavaScript. Be careful and make sure to not make your md files accessible for others though, as there are some security risks involved if using unsafe (Injections).

Make sure you have Go (v. 1.22+) installed and also make sure you have a Github Access Token (with public repo's set) for the API Pulls of your projects, aswell as your Resend API Key which is needed for the contact form ready.

1. **Create an about.md** It can be in any directory on your pc / server, just make sure to put the path in correctly in the following .env.
2. **Clone the Repository** ```git clone https://github.com/tekkpriest/go-portfolio-api.git```
3. **Create an .env file in your directory and add the following variables**  
    ```
    GITHUB_TOKEN=your_access_token  
    GITHUB_USERNAME=your_gh_username  
    RESEND_API_KEY=your_resend_api_key  
    EMAIL_FROM=noreply@yourdomain.com  
    EMAIL_TO=your_personal_email@example.com      
    ALLOWED=https://yourdomain.com    
    PORT=7302 (Default/Fallback, highly advised to use another port)
    ABOUT_MD_PATH=path_to_your_md_file
    ```
4. **Download Dependencies** ```go mod download```
5. **Build or run your server** ```go build -o go-portfolio-api ./cmd/server && ./go-portfolio-api``` or ```go run ./cmd/server/main.go```
