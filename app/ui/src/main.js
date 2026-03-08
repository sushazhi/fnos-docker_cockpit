import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import App from './App.vue'
import i18n from './i18n'
import './styles/main.css'

const routes = [
  { path: '/', name: 'home', component: () => import('./views/Home.vue') },
  { path: '/containers', name: 'containers', component: () => import('./views/Containers.vue') },
  { path: '/container/:id', name: 'container', component: () => import('./views/ContainerDetail.vue') },
  { path: '/images', name: 'images', component: () => import('./views/Images.vue') },
  { path: '/networks', name: 'networks', component: () => import('./views/Networks.vue') },
  { path: '/volumes', name: 'volumes', component: () => import('./views/Volumes.vue') },
  { path: '/compose', name: 'compose', component: () => import('./views/Compose.vue') },
  { path: '/compose/:name', name: 'compose-detail', component: () => import('./views/ComposeDetail.vue') },
  { path: '/system', name: 'system', component: () => import('./views/System.vue') },
  { path: '/settings', name: 'settings', component: () => import('./views/Settings.vue') }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

const app = createApp(App)
app.use(router)
app.use(i18n)
app.mount('#app')
