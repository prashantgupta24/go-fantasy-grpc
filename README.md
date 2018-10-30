# Go-fantasy
A Go-lang based fantasy premier league scraper!

It scraps off information from the official `fantasy premier league` site, and gets the player selection fraction from the current top 10 teams in the `Overall` league for all the gameweeks. The number denotes the number of teams the player is on.

> For example, in the output below, we see that `Wan-Bissaka` was in ALL of the top 10 teams of the `Overall` league at the end of `Gameweek 7`, whereas `Hennessey` was just in 3 of the top 10 teams.

## Output

![](https://github.com/prashantgupta24/go-fantasy/blob/master/output.jpg)

## Working

For each gameweek, I create a separate `go-routine` to fetch the player selection fraction for all the top 10 teams.

For example, after fetching the stats for `Gameweek 1`, I get:

| Player        | Gameweek 1           |
| ------------- |:-------------:|
| Wan-Bissaka     | 9 |
| Agüero      | 9      |
| Robertson | 8     |
| Mané | 8     |
| Mendy | 8     |

This states that `Wan-Bissaka` was selected in 9 of the top 10 teams, and so on.

## Comparison to single threaded application

The [single threaded](https://github.com/prashantgupta24/go-fantasy/tree/single-threaded) variation was the first iteration of the application, and it used to fetch each gameweek sequentially.

I ran both of them for 8 gameweeks (at the moment), and the results are already showing a difference:

#### Single threaded - Took 2.2s

```
Starting main program
Fetched data of 542 premier league players
Fetched 50 participants in league
Fetching data for gameweek  1
Fetching data for gameweek  2
Fetching data for gameweek  3
Fetching data for gameweek  4
Fetching data for gameweek  5
Fetching data for gameweek  6
Fetching data for gameweek  7
Fetching data for gameweek  8
Writing to file ...
Took 2.240394442s to fetch all data
```

#### Separate go-routines - Took 1.1s
```
Starting main program
Fetched data of 542 premier league players
Fetched 50 participants in league
Fetching data for gameweek  4
Fetching data for gameweek  9
Fetching data for gameweek  8
Fetching data for gameweek  1
Fetching data for gameweek  37
.
.
.
Data fetched for gameweek 1!
Data fetched for gameweek 5!
Data fetched for gameweek 4!
Data fetched for gameweek 3!
Data fetched for gameweek 8!
Data fetched for gameweek 7!
Data fetched for gameweek 6!
Data fetched for gameweek 2!
Writing to file ...
Took 1.131797205s to fetch all data!
```

**It's 2X as fast!**
