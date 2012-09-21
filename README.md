![My image](http://img.dafont.com/preview.php?text=hypeCMS&ttf=days0&ext=1&size=58&psize=s&y=57)

A CMS written in Go, using MongoDB as a backend. (Work in progress, heavily.)

Why the name
=======
Most people want his blog or web app to be hyped, right?

Installation
=======
- Install and start MongoDb
- go get github.com/opesun/hypecms
- Modify config values at the beginning of main.go if needed
- Clone github.com/opesun/hypecms-shared. Create a symbolic link in your file system named github.com/opesun/hypecms/shared which points to the cloned repo.
Alternatively, copy all contents of the hypecms-shared repo into the "shared" named folder in hypecms.

Stuff already working
=======
- Contents, custom content types, each with custom fields.
- Tags (categories).
- Content draft, versions.
- Cookie based user authentication.
- Online editing of template files. If deployed like a blogspot-like multiuser app: there are private templates, public templates, fork, publish features.
- Versioning of site state: any install/uninstall/configuration will create a newer version of site state, so you can revert to any previous version.
- Custom view editor: run any queries at any part of the page. The query builder generates excerpts, paging links etc.
- Any file is displayed as is, you dont even have to use any dynamic features of the CMS, you can simply copy-paste html pages and they will be displayed as if they were
dynamic content. Even in these html files, PHP-like require functionality is available.

Demo
=======
http://hypecms.com/

Current status
=======
Work in progress.

License
=======
Released under the 2-clause BSD license, see license.txt file.