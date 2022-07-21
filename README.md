# Trailhead Leaderboard API

A [Golang](https://go.dev) app that runs on [Heroku](https://heroku.com) to make callouts to `trailblazer.me/` and retrieves Trailblazer User data which can then be used in other applications. I recommend cloning this and running in your own Heroku instance.

> 🚨 Note: In order to retrieve data for a specific Trailhead User, they must have their profile set to public.

Trailhead changes their services occasionally so this an break at any time!

## Installation

Create a new Heroku app and push the source up. See the [Heroku docs](https://devcenter.heroku.com/articles/getting-started-with-go#deploy-the-app) for how to deploy the app.

```bash
$ heroku create
$ git push heroku master
$ heroku open
```

If you have Go [installed](https://golang.org/doc/install), you can run it locally by using the `run` command on `main.go`.

```bash
$ go run main.go
```

## Endpoints

This app has a few different endpoints for accessing public Trailhead data.

### Profile Data

```text
/trailblazer/matruff/profile
```

This endpoint returns public profile information found on trailblazer.me like "about me", company info, name, and profile/banner photos. [Example](https://go-trailhead-leaderboard-api.herokuapp.com/trailblazer/matruff/profile)

### Badge Data

This endpoint returns badges earned by the Trailblazer. The basic call without any options only gives a max of 8 at a time, representing your most recently earned badges.

```text
/trailblazer/matruff/badges
```

You can also filter badges by sending a string at the end of the URL without or without optional count. Possible filter values are: all, module, superbadge, event, and project. [Example](https://go-trailhead-leaderboard-api.herokuapp.com/trailblazer/matruff/badges/superbadge)

```text
/trailblazer/matruff/badges/superbadge
```

or

```text
/trailblazer/matruff/badges/all/24
```

The GraphQl endpoint returns a `pageInfo` record in the response that you can use to query for additional pages past the first count you provide. For example, in your first response for a coutn of 16 you'll get the `endCursor` value to use in your next callout to get the page after it. Pass that at the end of the API endpoint and you'll get the next 16 results. The response also tells you if there is a next page.

`pageInfo` response example:

```json
"pageInfo": {
    "__typename": "PageInfo",
    "endCursor": "eyJzIjoiMDI2NmQzOGEtZjc1MS0zOTEwLTQ0NGItOW[...]",
    "hasNextPage": true,
    "startCursor": "eyJzIjoiMDI5YWQzNjItYWQ0Mi0yOTY0LWNlMTMt[...]",
    "hasPreviousPage": false
}
```

Using `endCursor` as `after`:

```text
/trailblazer/matruff/badges/all/16/24eyJzIjoiMDI2NmQzOGEtZjc1MS0aOTEwLItOWU4MzRx[...]
```

### Certifications Data

```text
/trailblazer/matruff/certifications
```

This endpoint returns Certifications the Trailblazer has achieved. [Example](https://go-trailhead-leaderboard-api.herokuapp.com/trailblazer/matruff/certifications)

## Special Thanks

Thanks to both [@Patlatus](https://github.com/Patlatus/Salesforce-Trailhead-Api-Hack) and [@krankekatze](https://github.com/krankekatze/trailhead-batch) for the inspiration to build this. Check out their repos for related solutions.
