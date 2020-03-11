# Go Trailhead Leaderboard API

A Golang app that runs on Heroku to make callouts to `trailhead.me/` and displays returned JSON data. That data can then be used in other applications. You can also clone and deploy this to your own Heroku instance.

> Note that in order to retrieve data for a specific Trailhead User, they must have their profile set to public.

## Installation

Create a new Heroku app and push the source up. See the [Heroku docs](https://devcenter.heroku.com/articles/getting-started-with-go#deploy-the-app) for how to deploy the app.

``` bash
$ heroku create
$ git push heroku master
$ heroku open
```

If you have [Go](https://golang.org/doc/install) installed, you can run it locally by using the `run` command on `main.go`.

``` bash
$ go run main.go
```

## Use

This app has a few different endpoints for accessing public Trailhead data.

### Profile Data

``` text
https://go-trailhead-leaderboard-api.herokuapp.com/trailblazer/<trailhead_handle>
```

This endpoint returns information about skills learned, counts of badges, points, trails, points til next rank, and rank. [Example](https://go-trailhead-leaderboard-api.herokuapp.com/trailblazer/matruff)

``` text
https://go-trailhead-leaderboard-api.herokuapp.com/trailblazer/<trailhead_handle>/profile
```

This endpoint returns public Profile information found on trailblazer.me like about me data, company info, name, and profile/banner photo. [Example](https://go-trailhead-leaderboard-api.herokuapp.com/trailblazer/matruff/profile)

### Badge Data

``` text
https://go-trailhead-leaderboard-api.herokuapp.com/trailblazer/<trailhead_handle>/badges
```

This endpoint returns badges earned by the Trailblazer. The API only gives a max of 30 at a time. You can filter badges by sending a string at the end of the URL. Possible filter values are: all, module, superbadge, event, and project. [Example](https://go-trailhead-leaderboard-api.herokuapp.com/trailblazer/matruff/badges/superbadge)

``` text
https://go-trailhead-leaderboard-api.herokuapp.com/trailblazer/<trailhead_handle>/badges/<offset>
```

Add an offset to the end of the filter will allow you to offest your badge query by that many. For example, because the API only returns 30 at a time, `/badges/module` will only get you the 30 most recent badges of type module. To get the next 30 you would call `/badges/module/30` to offset the query and return the next 30 badges. [Example](https://go-trailhead-leaderboard-api.herokuapp.com/trailblazer/matruff/badges/module/30)
