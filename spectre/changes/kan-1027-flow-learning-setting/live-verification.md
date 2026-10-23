# Live verification — kan-1027-flow-learning-setting

Against the change's own workspace store (flow_kan_1027_5c4f, flowd on 127.0.0.1:4493), never the dev flowd.

- before (merge base c1282bc4, detached worktree): `flow settings get` → `{"reviewers":["primary","mutation"]}` — no learning field; `flow learning list` → `unknown command "learning"`.
- after: `get` on the row written before 0043 → `"learning":true`; `set -learning off` then `get` → `{"reviewers":["primary","mutation"],"learning":false}` — reviewers unchanged; `learning add live-check-topic` ×3 exit 0, 4th exit 1 naming `already shown three times`; `learning list` → `[{"topic":"live-check-topic","count":3,"lastNote":"x"}]`.
- failure look (learning missing, reviewers reset, fourth add accepted): not seen — matched.
