"""Outbound Slack tools. The bot token stays here, in the worker — Norns
never sees it."""

import httpx

from norns import tool

from .. import config


def _slack_api(method: str, payload: dict) -> dict:
    resp = httpx.post(
        f"https://slack.com/api/{method}",
        json=payload,
        headers={"Authorization": f"Bearer {config.SLACK_BOT_TOKEN}"},
        timeout=15,
    )
    resp.raise_for_status()
    data = resp.json()
    if not data.get("ok"):
        raise RuntimeError(f"Slack {method} failed: {data.get('error', 'unknown error')}")
    return data


@tool(side_effect=True)
def post_to_slack(channel: str, text: str) -> str:
    """Post a message to a Slack channel. Channel is a name like #general or a channel ID."""
    data = _slack_api("chat.postMessage", {"channel": channel, "text": text})
    return f"Posted to {channel} (ts {data['ts']})"


@tool
def list_slack_channels() -> str:
    """List the public Slack channels this bot can see, as 'name: id' lines."""
    data = _slack_api("conversations.list", {"limit": 100, "exclude_archived": True})
    lines = [f"#{c['name']}: {c['id']}" for c in data.get("channels", [])]
    return "\n".join(lines) or "No channels visible to this bot."
