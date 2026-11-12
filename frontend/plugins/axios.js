import axios from 'axios';

export default defineNuxtPlugin(() => {
  axios.defaults.baseURL = 'http://localhost:8080'; // Замените на ваш сервер
  axios.defaults.withCredentials = true; // Для отправки токенов через cookies
});
