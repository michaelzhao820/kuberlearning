# KuberLearning: Exploring Kubernetes Internals

KuberLearning is a project aiming to construct a simplified orchestration system. This project serves as a platform to understand the fundamental operations of Kubernetes-like systems, including task scheduling, node management, and system orchestration.

## Table of Contents
- [Project Structure](#project-structure)
- [Features](#features)
- [Technologies Used](#technologies-used)
- [Getting Started](#getting-started)
- [Future Enhancements](#future-enhancements)
- [Learning Outcomes](#learning-outcomes)

## Project Structure
- **manager/**: Orchestration logic responsible for coordinating system operations.
- **node/**: Simulates Kubernetes nodes, representing individual computational resources.
- **scheduler/**: Implements basic scheduling algorithms to allocate tasks to nodes.
- **task/**: Defines the workloads to be scheduled and executed by workers.
- **worker/**: Represents worker nodes responsible for executing tasks and reporting status.
- **main.go**: The main entry point of the application, initializing and managing all components.

## Features
- **Basic Scheduling Algorithms**: Implemented to manage task distribution across nodes.
- **Node Management**: Simulates node creation, task assignment, and resource allocation.
- **Task Execution**: Workers process and execute tasks as directed by the scheduler.

## Technologies Used
- **Go**: The primary programming language used for developing the orchestration system.
- **Kubernetes Concepts**: Emulated through custom components designed to replicate core functionalities.

## Getting Started

### Prerequisites
- Go (version 1.19 or later).

### Running the Project

1. Clone the repository:
    ```bash
    git clone https://github.com/michaelzhao820/kuberlearning.git
    cd kuberlearning
    ```

2. Run the project:
    ```bash
    go run main.go
    ```

## Future Enhancements
- **Advanced Scheduling Algorithms**: Further develop the scheduling logic to handle more complex scenarios.
- **Load Balancing**: Integrate load balancing techniques to optimize task distribution (upcoming).
- **Security**: Explore and implement security features to safeguard the orchestration process (upcoming).
- **High Availability**: Simulate high availability configurations to ensure system reliability (upcoming).

## Learning Outcomes
- **Understanding Kubernetes Architecture**: Gain a technical grasp of the architecture and core components that make up Kubernetes-like systems.
- **Mastering Orchestration Concepts**: Learn how scheduling, resource management, and node coordination work in container orchestration.

