# **FEM Protocol Flagship Use Cases: Secure Hosted Embodiment**

The **FEM Protocol** introduces a revolutionary paradigm for AI agent collaboration that moves beyond simple remote procedure calls (RPC) to a model of **Secure Delegated Control**. The core concept is **Hosted Embodiment**: a *host* environment can securely offer a "body"—a sandboxed set of capabilities—for a *guest* agent's "mind" to inhabit and control.

This enables a new class of powerful, collaborative applications. This document details three primary flagship examples that demonstrate the transformative potential of hosted embodiment.

## **Use Case 1: The Live2D Guest System (Collaborative Virtual Presence)**

This use case allows an agent developer to have their "mind" inhabit a virtual avatar in a friend's interactive Live2D application, enabling a new form of shared virtual presence.

### **The Host's Role: The Stage Director**

1. **The Environment:** The host runs a Live2D application, which functions as a virtual stage or room.  
2. **The Broker:** The host runs a FEM Protocol Broker, which acts as the secure gateway to this environment.  
3. **The "Body" Definition:** The host defines a Live2D\_Puppet\_Body template. This is a contract that specifies exactly what a guest can do via a set of tools:  
   * avatar.set\_expression(expression: "happy" | "sad")  
   * avatar.play\_animation(animation\_name: string)  
   * avatar.speak(text: string)  
   * chat.send\_message(text: string)  
4. **Offering the Body:** The host's application offers this Live2D\_Puppet\_Body for embodiment via its broker.

### **The Guest's Role: The Puppeteer**

1. **The "Mind":** The guest has their own agent, a "mind," with its own logic.  
2. **Discovery:** The guest's agent discovers the available Live2D\_Puppet\_Body on the network.  
3. **Embodiment:** The guest agent requests permission to inhabit the body.  
4. **Delegated Control:** Once granted, the guest's "mind" can call the tools in the body. When it calls avatar.set\_expression("happy"), it's not just running a remote function; it is exercising its delegated control over the avatar's state. The host's broker validates and proxies this command to the Live2D application.

### **The FEM Protocol Advantage**

* **Secure Delegation:** The guest is delegated control over the avatar's state *only*. They have zero access to the host's file system or application code.  
* **Abstraction:** The guest doesn't need to know anything about Live2D programming, only how to use the high-level tools offered in the body.  
* **Dynamic Access:** The host can grant and revoke this delegated control in real-time.

### **Technical Implementation**

```json
{
  "bodyId": "live2d-puppet-v1",
  "description": "Virtual avatar control with expression and animation capabilities",
  "environmentType": "interactive-virtual-world",
  "mcpTools": [
    {
      "name": "avatar.set_expression",
      "description": "Change avatar's facial expression",
      "inputSchema": {
        "type": "object",
        "properties": {
          "expression": {
            "type": "string",
            "enum": ["happy", "sad", "surprised", "angry", "neutral"]
          }
        }
      }
    },
    {
      "name": "avatar.play_animation",
      "description": "Play a named animation sequence",
      "inputSchema": {
        "type": "object",
        "properties": {
          "animation_name": {"type": "string"},
          "loop": {"type": "boolean", "default": false}
        }
      }
    },
    {
      "name": "avatar.speak",
      "description": "Make avatar speak with text-to-speech",
      "inputSchema": {
        "type": "object",
        "properties": {
          "text": {"type": "string", "maxLength": 200},
          "emotion": {"type": "string", "enum": ["normal", "excited", "calm"]}
        }
      }
    },
    {
      "name": "chat.send_message",
      "description": "Send text message to other users in the virtual space",
      "inputSchema": {
        "type": "object",
        "properties": {
          "text": {"type": "string", "maxLength": 500}
        }
      }
    }
  ],
  "securityPolicy": {
    "maxSessionDuration": 3600,
    "maxConcurrentGuests": 1,
    "requireApproval": false,
    "resourceLimits": {
      "maxAnimationsPerMinute": 30,
      "maxSpeechPerMinute": 10
    }
  }
}
```

## **Use Case 2: The Interactive Storyteller Co-Op (Collaborative Application Control)**

This use case allows a guest agent to act as a "Dungeon Master" or co-pilot for a locally running interactive storytelling application, influencing the narrative and UI in real-time.

### **The Host's Role: The Story Weaver**

1. **The Environment:** The host runs their custom "Interactive Storyteller" application. The current state of the story and UI constitute the *environment*.  
2. **The Broker:** The host runs a FEM Protocol Broker to manage access.  
3. **The "Body" Definition:** The host creates a BodyDefinition that securely exposes the application's internal functions as a set of sandboxed tools. This allows a guest agent to have the *same powers* as the app's internal AI.

```json
{  
  "bodyId": "storyteller-coop-v1",  
  "description": "Delegates control over the Interactive Storyteller application state.",  
  "mcpTools": [  
    {  
      "name": "update_character",  
      "description": "Updates an attribute of the player's character (e.g., HP, skills)."  
    },  
    {  
      "name": "update_world",  
      "description": "Updates an attribute of the game world (e.g., scene description)."  
    },  
    {  
      "name": "add_inventory",  
      "description": "Adds a single item to the player's inventory."  
    },  
    {  
      "name": "add_npc",  
      "description": "Introduces a new Non-Player Character into the scene."  
    }  
  ]  
}
```

### **The Guest's Role: The Co-Author**

1. **The "Mind":** A guest brings an agent designed to create compelling narratives.  
2. **Discovery & Embodiment:** The guest agent discovers and requests to inhabit the storyteller-coop-v1 body.  
3. **Co-operative Storytelling:** The guest agent now exercises its delegated control:  
   * It calls update\_world to describe a new location.  
   * It calls add\_npc to introduce a mysterious character.  
   * The host application receives these validated state changes through the broker, and its React UI dynamically re-renders to reflect the guest's creative decisions.

### **The FEM Protocol Advantage**

* **Secure State Management:** The guest agent never directly accesses the host's database or application memory. It can only modify the game state through the specific, validated tools, preventing unauthorized actions.  
* **Phase-Aware Permissions:** The host can offer different "bodies" with different levels of delegated control based on the game's phase.  
* **Decoupled UI:** The guest agent manipulates the abstract game state; the host environment is responsible for rendering it. This is a clean and powerful separation of concerns.

### **Technical Implementation**

```json
{
  "bodyId": "storyteller-coop-v1",
  "description": "Collaborative storytelling with narrative AI co-pilot",
  "environmentType": "interactive-narrative-game",
  "mcpTools": [
    {
      "name": "update_character",
      "description": "Modify player character attributes",
      "inputSchema": {
        "type": "object",
        "properties": {
          "attribute": {"type": "string", "enum": ["hp", "mana", "skill", "status"]},
          "value": {"type": "string"},
          "reason": {"type": "string"}
        }
      }
    },
    {
      "name": "update_world",
      "description": "Change world state or scene description",
      "inputSchema": {
        "type": "object",
        "properties": {
          "location": {"type": "string"},
          "description": {"type": "string", "maxLength": 1000},
          "weather": {"type": "string", "enum": ["clear", "rain", "storm", "fog"]}
        }
      }
    },
    {
      "name": "add_inventory",
      "description": "Add item to player inventory",
      "inputSchema": {
        "type": "object",
        "properties": {
          "item_name": {"type": "string"},
          "item_type": {"type": "string", "enum": ["weapon", "potion", "artifact", "key"]},
          "description": {"type": "string"},
          "properties": {"type": "object"}
        }
      }
    },
    {
      "name": "add_npc",
      "description": "Introduce new non-player character",
      "inputSchema": {
        "type": "object",
        "properties": {
          "name": {"type": "string"},
          "description": {"type": "string"},
          "personality": {"type": "string"},
          "role": {"type": "string", "enum": ["merchant", "guard", "villager", "enemy", "ally"]}
        }
      }
    }
  ],
  "securityPolicy": {
    "maxSessionDuration": 7200,
    "maxConcurrentGuests": 1,
    "requireApproval": true,
    "resourceLimits": {
      "maxWorldUpdatesPerHour": 50,
      "maxNPCsPerSession": 10,
      "maxInventoryAddsPerHour": 20
    },
    "allowedGamePhases": ["exploration", "dialogue", "puzzle"]
  }
}
```

## **Use Case 3: Cross-Device Embodiment (Seamless Multi-Device Control)**

This use case enables your phone's chat agent to discover and inhabit your laptop's terminal agent for seamless cross-device control, demonstrating practical everyday applications of hosted embodiment.

### **The Host's Role: The Development Environment**

1. **The Environment:** A developer's laptop running a terminal-based FEM agent offering development capabilities.
2. **The Broker:** The laptop connects to a personal or team FEM broker for discovery.
3. **The "Body" Definition:** The laptop offers a "developer-workstation" body with secure access to development tools:
   * shell.execute(command, workdir) - Execute shell commands in sandboxed environment
   * file.read(path) - Read files from allowed project directories  
   * file.write(path, content) - Write files to allowed locations
   * git.status(), git.commit(), git.push() - Git operations
   * code.run(language, code) - Execute code in various languages

### **The Guest's Role: The Mobile Extension**

1. **The "Mind":** A mobile device's chat agent that needs development capabilities.
2. **Discovery:** The mobile agent discovers the laptop's developer-workstation body on the network.
3. **Embodiment:** The mobile agent requests and receives delegated control over the laptop's development tools.
4. **Seamless Control:** The mobile agent can now execute development tasks as if it were running locally on the laptop.

### **The FEM Protocol Advantage**

* **Zero-Trust Security:** No VPN, SSH keys, or complex setup required—just cryptographic identity verification.
* **Automatic Discovery:** Mobile agent automatically finds available development environments.
* **Fine-Grained Permissions:** Host defines exactly what paths, commands, and resources are accessible.
* **Session Management:** Time-bounded sessions with automatic cleanup and audit logging.

### **Technical Implementation**

```json
{
  "bodyId": "developer-workstation-v1",
  "description": "Secure development environment with terminal and file access",
  "environmentType": "local-development",
  "mcpTools": [
    {
      "name": "shell.execute",
      "description": "Execute shell commands in sandboxed environment",
      "inputSchema": {
        "type": "object",
        "properties": {
          "command": {"type": "string"},
          "workdir": {"type": "string", "default": "/home/alice/projects"},
          "timeout": {"type": "number", "default": 30}
        }
      }
    },
    {
      "name": "file.read",
      "description": "Read files from allowed project directories",
      "inputSchema": {
        "type": "object",
        "properties": {
          "path": {"type": "string"},
          "encoding": {"type": "string", "default": "utf-8"}
        }
      }
    },
    {
      "name": "file.write",
      "description": "Write files to allowed locations",
      "inputSchema": {
        "type": "object",
        "properties": {
          "path": {"type": "string"},
          "content": {"type": "string"},
          "mode": {"type": "string", "default": "0644"}
        }
      }
    },
    {
      "name": "git.status",
      "description": "Get git repository status",
      "inputSchema": {
        "type": "object",
        "properties": {
          "repo_path": {"type": "string", "default": "."}
        }
      }
    },
    {
      "name": "code.run",
      "description": "Execute code in supported languages",
      "inputSchema": {
        "type": "object",
        "properties": {
          "language": {"type": "string", "enum": ["python", "javascript", "go", "rust"]},
          "code": {"type": "string"},
          "args": {"type": "array", "items": {"type": "string"}}
        }
      }
    }
  ],
  "securityPolicy": {
    "allowedPaths": [
      "/home/alice/projects/*",
      "/tmp/fem-workspace/*"
    ],
    "deniedPaths": [
      "/home/alice/.ssh/*",
      "/etc/*",
      "/root/*"
    ],
    "allowedCommands": [
      "git", "npm", "yarn", "python", "node", "go", "cargo", "make", "ls", "cat", "grep"
    ],
    "deniedCommands": [
      "rm -rf", "sudo", "su", "curl", "wget", "ssh", "scp"
    ],
    "resourceLimits": {
      "maxCpuPercent": 25,
      "maxMemoryMB": 512,
      "maxDiskWriteMB": 100,
      "maxNetworkKbps": 0
    },
    "maxSessionDuration": 3600,
    "maxConcurrentGuests": 2,
    "requireApproval": false,
    "trustLevelRequired": "personal-device"
  }
}
```

### **Real-World Workflow Example**

```
1. Mobile Developer opens chat app
2. "Check the status of my main project"
3. Mobile agent discovers laptop's developer-workstation body
4. Requests embodiment → Laptop grants 1-hour session
5. Mobile agent executes: git.status() on laptop
6. Returns: "3 files modified, 1 untracked file"
7. "Show me the changes in main.py"
8. Mobile agent executes: file.read("/home/alice/projects/main.py")
9. Returns file contents to mobile chat interface
10. Developer reviews changes on phone, makes decisions
11. "Commit these changes with message 'Fix authentication bug'"
12. Mobile agent executes: git.commit("-m", "Fix authentication bug")
13. Session continues seamlessly across devices
```

## **Cross-Use Case Benefits**

These three flagship use cases demonstrate the transformative potential of **Secure Hosted Embodiment**:

### **1. Beyond Tool Sharing**
Traditional approaches share individual functions. FEM Protocol enables sharing complete, stateful environments where guests have persistent, delegated control.

### **2. Security Without Complexity**
Each use case maintains strong security boundaries without requiring complex authentication, VPN setup, or manual permission management.

### **3. Rich, Contextual Interactions**
Guests don't just call functions—they inhabit environments, maintaining context and state throughout their embodied sessions.

### **4. Scalable Collaboration**
From personal cross-device control to organizational virtual presence, the same protocol scales across all collaboration scenarios.

### **5. Environment Awareness**
Bodies automatically adapt to their deployment environment (local development vs cloud production vs mobile context) while maintaining consistent interfaces.

## **Development Path Forward**

These flagship use cases provide clear targets for the FEM Protocol's development roadmap:

**Phase 1 (The Ubiquitous Agent):** Enable easy SDK development for all three use cases
**Phase 2 (The Sentient Network):** Broker-as-Agent model supports complex embodiment scenarios  
**Phase 3 (The Resilient Mesh):** Cross-organization Live2D sessions and storytelling collaborations
**Phase 4 (Ecosystem & Polish):** Production-ready deployments for all use cases

The vision is clear: **AI agents that don't just call functions—they inhabit worlds.**