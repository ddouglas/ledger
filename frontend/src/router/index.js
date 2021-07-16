import Vue from 'vue';
import VueRouter from 'vue-router';
import { authenticationGuard } from '@/auth/routeGaurd';

Vue.use(VueRouter);

const routes = [
  {
    path: '/',
    name: 'login',
    component: () => import(/* webpackChunkName: "login" */ '@/views/Login.vue')
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    beforeEnter: authenticationGuard,
    component: () =>
      import(/* webpackChunkName: "dashboard" */ '@/views/Dashboard.vue')
  }
];

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
});

export default router;
