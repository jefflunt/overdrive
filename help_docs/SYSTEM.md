[← Back to Help Home](/help)

# System Settings

The System Settings screen provides information about the overall health of the overdrive system and access to administrative controls.

## System Health

The health section provides real-time monitoring of the overdrive environment:

*   **API Uptime**: Indicates how long the overdrive API server has been running since its last start.
*   **Scheduler Uptime**: Indicates how long the background job scheduler has been active.
*   **Active Containers**: Displays the total number of currently running job, chat, or command containers across all projects.

## Appearance and Themes

overdrive features a flexible theme system to customize your workspace. You can change the color scheme by clicking on your user profile icon in the sidebar and selecting the **Theme** button.

### Built-in Themes

We offer several pre-configured themes, including:

*   **Dark Themes**: Dark (Default), Dracula, Monokai, Nord, Solarized Dark, Gruvbox Dark, One Dark, Tokyo Night, Synthwave, and Cyberpunk.

### Custom Themes

For full control over your environment's appearance, you can create a **Custom Theme**. This allows you to individually set colors for the background, panels, borders, primary accents, and text. You can also toggle between light and dark modes for your custom configuration.

## Danger Zone

The **Danger Zone** contains powerful actions that should be used with caution, as they temporarily interrupt service availability:

*   **Rebuild and Restart API**: Recompiles the Go API server and restarts the service.
*   **Rebuild and Restart Scheduler**: Recompiles the job scheduler, stops all running containers, and restarts the background loop.

These actions are typically used for system updates or to resolve persistent service issues.
