import { defineNuxtRouteMiddleware, useState, navigateTo } from '#imports'

export default defineNuxtRouteMiddleware((to) => {
  // Получаем информацию о пользователе из состояния приложения
  const user = useState('user').value
  
  // Проверяем наличие пользователя и защищенность пути
  if (!user && to.path.includes('/dashboard')) {
    console.log('Middleware: пользователь не аутентифицирован, перенаправление на /login')
    // Сохраняем текущий путь для возможного редиректа после входа
    useState('redirectPath', () => to.fullPath)
    // Перенаправляем на страницу входа
    return navigateTo('/login')
  }
  
  // Если пользователь авторизован, но пытается получить доступ 
  // к странице для другой роли
  if (user && to.path.includes('/dashboard/')) {
    const role = user.role
    const requestedRole = to.path.split('/dashboard/')[1].split('/')[0]
    
    // Проверяем соответствие роли и URL
    if (requestedRole !== role) {
      console.log(`Middleware: несоответствие ролей, перенаправление на /dashboard/${role}`)
      // Если роль не совпадает, перенаправляем на страницу с правильной ролью
      return navigateTo(`/dashboard/${role}`)
    }
  }
}) 