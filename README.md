# disport

A simple discord bot to report discord users to the admin and send him actionable insights based on promcord data.

## Why

Discord Communities like most active communications channels online have been subject to trolls, spam and misbehaving people since the day of their birth.
While active moderation and a well minded community can actively engage such bad behavior, this project is part of a chain of projects to allow data driven moderation as well as proactive reactions.

Based on the generated insights by projects like `promcord` we aim to allow admins making quicker decisions based on community requests (users reporting other user) founded by real data insights.

`disport` itself is the bot allowing users to report others and then sends this report to the admin(s) containing actionable insights on the reported user from `promcord`.

## How

This service should connect to a server and listen to every message in all channels on it.
For starters, we only offer one action `report`.

Calling the bot by mention in the form of `@disport <user mention>` will start the analysis inside `disport` collecting all information on the user we have. For starters this is:

* User Message Rate (`msg_count` rate from promcord)
* User Bad Word Count
* Average Message Length

In a later stage we would like to get more specialized information like:

* Sentiment Analysis
* User Rating on different servers
* Possibly we could even try to identify users abusing new accounts to access servers again by doing various recognition concepts (text style, online times, etc.)

## Development

This project is built using [Bazel](https://bazel.build).
To build and run the code directly, install Bazel and run `make run CMD=//cmd/disport`.

We build directly into a docker image and deploy it to Kubernetes.
The image step can be triggered manually by running `make docker CMD=//cmd/disport`.

To deploy directly to Kubernetes, run `make kube CMD=//cmd/disport:dev` which we use internally as dev deployment target.
There will be a `:prod` target in the future.

## Coding and Style

Our code is always checked by Travis using `make test check` therefor all Golang rules on syntax and formating have to be met for pull requests to be merged.
While this might incur more work for possible contributors, we see the code produced here as production critical once finished and therefor strive for high code quality.

We are developing this mostly using TDD and BDD. If you don't know what this is, we recommend this [video](https://www.youtube.com/watch?v=uFXfTXSSt4I) for starters.

Please do reasonable commit sizes.


## Dependencies
All dependencies inside this project are being managed by [dep](https://github.com/golang/dep) and are checked in.
After pulling the repository, it should not be required to do any further preparations aside from `make deps` to prepare the dev tools (once).

If new dependencies get added while coding, make sure to add them using `dep ensure --add "importpath"` and to check them into git.
We recommend adding your vendor changes in a separate commit to make reviewing your changes easier and faster.

## Testing
To run tests you can use:
```bash
make test
```

## Contributing

Feedback and contributions are highly welcome. Feel free to file issues, feature or pull requests.
If you are interested in using this project now or in a later stage, feel free to get in touch.
If you are developing or already finished solutions that rely on rcon in any way, we would be happy to talk to you for both gaining insights as well as looking for options to collaborate and involve your project into other endeavors.

We are always looking for active contributors, team members and partner projects sharing our vision.
Easiest way of reaching us is via [Discord](https://discord.gg/dWZkR6R).

See you soon,
the PlayNet Team

## Attributions

* [Kolide for providing `kit`](https://github.com/kolide/kit)

## License

This project's license is located in the [LICENSE file](LICENSE).