# Qiscus Agent Allocator – RadiantSkin

Custom agent allocation system for Qiscus Omnichannel platform, built to automate agent assignment using FIFO logic, online status, and configurable capacity per agent — helping RadiantSkin handle thousands of daily inquiries efficiently.

---


## 📋 Submission Checklist
1. **App ID**: `qsgwv-xfuefkhy5ixjqnt`  
2. **Git Repo**: `https://github.com/alghoziii/qisqus-automation-layer-`  
3. **Live Demo**: `https://qiscus-app-production.up.railway.app/webhook/agent_allocation`  

---

## 🚀 Features

- ✅ Automatic assignment of customers to agents
- ✅ Only assigns **online** agents
- ✅ Respects **maximum active customer limit** (default: 2 per agent)
- ✅ **FIFO queue** system if no agent is available
- ✅ Built with **Golang** using `Gin`, `GORM`, and `PostgreSQL`

---

## 📬 Webhook Custom Agent Allocation
`https://qiscus-app-production.up.railway.app/webhook/agent_allocation`


## 🗝️ Testing Mark As Resolved LocalPostman

`https://qiscus-app-production.up.railway.app/resolve`


body json
`{
  "room_id": "328386524"
}
`

## 💻 How to Run the Project Locally

## Clone Repositori
- ✅ git@github.com:alghoziii/qisqus-automation-layer-.git
- ✅ cd qiscus-automation-layer 

## Install Dependency Go dan jalankan Aplikasi
- ✅ go mod tidy
- ✅ go run cmd/main.go

## Setup Docker
- ✅ docker build -t ozzyyyy/qiscus-app:latest .
- ✅ docker push ozzyyyy/qiscus-app:latest   


