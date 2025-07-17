<a id="readme-top"></a>
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![Unlicense License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]



<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/DanielProtopopov/hideout">
    <img src="data/images/logo.png" alt="Logo" width="80" height="80">
  </a>

<h3 align="center">Hideout</h3>

  <p align="center">
    A secrets manager that provides an easy-to-use API (instead of GUI)
    <br />
    <a href="https://github.com/DanielProtopopov/hideout"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/DanielProtopopov/hideout">View Demo</a>
    &middot;
    <a href="https://github.com/DanielProtopopov/hideout/issues/new?labels=bug&template=bug-report---.md">Report Bug</a>
    &middot;
    <a href="https://github.com/DanielProtopopov/hideout/issues/new?labels=enhancement&template=feature-request---.md">Request Feature</a>
  </p>
</div>



<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project

[![Product Name Screen Shot][product-screenshot]](https://github.com/DanielProtopopov/hideout)

There are many great secret managers out there; however, I didn't find one that really suited my needs so I created this enhanced one.
I want to create a secrets manager that is used mostly via API, contain copy-and-paste capabilities across "folders", include reference links and tokenized (scriptable) values.

Here's why:
* Your time should be focused on managing common secrets, operating with references in other places
* You shouldn't be doing the same tasks over and over via GUI because it is very time-consuming

Of course, no one secrets manager will serve all projects since your needs may be different. So I'll be adding more in the near future. You may also suggest changes by forking this repo and creating a pull request or opening an issue. Thanks to all the people have contributed to expanding it!

<p align="right">(<a href="#readme-top">back to top</a>)</p>



### Built With


* [![Golang][Golang]][Golang-url]
* [![go-i18n][go-i18n]][go-i18n-url]
* [![gofakeit][gofakeit]][gofakeit-url]
* [![swaggo-swag][swaggo-swag]][swaggo-swag-url]

* [![Docker][Docker]][Docker-url]
* [![Taskfile][Taskfile]][Taskfile-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- GETTING STARTED -->
## Getting Started

To get a local copy up and running follow these simple example steps.

### Prerequisites

This is an example of how to list things you need to use the software and how to install them.
* npm
  ```sh
  npm install npm@latest -g
  ```

### Installation

_Below is an example of how you can instruct your audience on installing and setting up your app. This template doesn't rely on any external dependencies or services._

1. Get a free API Key at [https://github.com/DanielProtopopov/hideout](https://github.com/DanielProtopopov/hideout)
2. Clone the repo
   ```sh
   git clone https://github.com/github_username/repo_name.git
   ```
3. Install NPM packages
   ```sh
   npm install
   ```
4. Enter your API in `config.js`
   ```js
   const API_KEY = 'ENTER YOUR API';
   ```
5. Change git remote url to avoid accidental pushes to base project
   ```sh
   git remote set-url origin github_username/repo_name
   git remote -v # confirm the changes
   ```

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- USAGE EXAMPLES -->
## Usage

Use this space to show useful examples of how a project can be used. Additional screenshots, code examples and demos work well in this space. You may also link to more resources.

_For more examples, please refer to the [Documentation](https://github.com/DanielProtopopov/hideout)_

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ROADMAP -->
## Roadmap

- [ ] Add file storage adapter
- [ ] Add virtual filesystem adapter
- [ ] Add redis storage adapter
- [ ] Add database storage adapter
- [ ] Add references (linking) mechanism for secrets
- [ ] Add copy-paste secrets mechanism
- [ ] Add authentication mechanism
- [ ] Add dynamic secrets via risor-io
- [ ] Add access control mechanisms via Casbin

See the [open issues](https://github.com/DanielProtopopov/hideout/issues) for a full list of proposed features (and known issues).

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Top contributors:

<a href="https://github.com/DanielProtopopov/hideout/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=DanielProtopopov/hideout" alt="contrib.rocks image" />
</a>

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- LICENSE -->
## License

Distributed under the Unlicense License. See `LICENSE.txt` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTACT -->
## Contact

Your Name - [@danielprotopopov](https://x.com/danielprotopopov) - danielprotopopov@gmail.com

Project Link: [https://github.com/DanielProtopopov/hideout](https://github.com/DanielProtopopov/hideout)

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

Use this space to list resources you find helpful and would like to give credit to. I've included a few of my favorites to kick things off!

* [Choose an Open Source License](https://choosealicense.com)
* [GitHub Emoji Cheat Sheet](https://www.webpagefx.com/tools/emoji-cheat-sheet)
* [GitHub Pages](https://pages.github.com)

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/othneildrew/Best-README-Template.svg?style=for-the-badge
[contributors-url]: https://github.com/DanielProtopopov/hideout/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/othneildrew/Best-README-Template.svg?style=for-the-badge
[forks-url]: https://github.com/DanielProtopopov/hideout/network/members
[stars-shield]: https://img.shields.io/github/stars/othneildrew/Best-README-Template.svg?style=for-the-badge
[stars-url]: https://github.com/DanielProtopopov/hideout/stargazers
[issues-shield]: https://img.shields.io/github/issues/othneildrew/Best-README-Template.svg?style=for-the-badge
[issues-url]: https://github.com/DanielProtopopov/hideout/issues
[license-shield]: https://img.shields.io/github/license/othneildrew/Best-README-Template.svg?style=for-the-badge
[license-url]: https://github.com/DanielProtopopov/hideout/blob/master/LICENSE.txt
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://linkedin.com/in/danielprotopopov
[Golang]: https://img.shields.io/badge/golang-000000?style=for-the-badge&logo=golang&logoColor=white
[Golang-url]: https://go.dev/
[go-i18n]: https://img.shields.io/badge/golang-000000?style=for-the-badge&logo=golang&logoColor=white
[go-i18n-url]: https://github.com/nicksnyder/go-i18n
[gofakeit]: https://img.shields.io/badge/golang-000000?style=for-the-badge&logo=golang&logoColor=white
[gofakeit-url]: https://github.com/brianvoe/gofakeit
[swaggo-swag]: https://img.shields.io/badge/golang-000000?style=for-the-badge&logo=golang&logoColor=white
[swaggo-swag-url]: https://github.com/swaggo/swag
[Docker]: https://img.shields.io/badge/docker-000000?style=for-the-badge&logo=docker&logoColor=white
[Docker-url]: https://www.docker.com/
[Taskfile]: https://img.shields.io/badge/golang-000000?style=for-the-badge&logo=golang&logoColor=white
[Taskfile-url]: https://github.com/go-task/task