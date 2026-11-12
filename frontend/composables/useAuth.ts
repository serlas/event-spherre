import { ref, computed } from 'vue'
import { useState, navigateTo, useRuntimeConfig } from '#imports'

export function useAuth() {
  // Загружаем пользователя из localStorage при инициализации
  const loadUserFromStorage = () => {
    if (process.client) {
      try {
        const savedUser = localStorage.getItem('user')
        if (savedUser) {
          const userData = JSON.parse(savedUser)
          // При загрузке из localStorage сразу запрашиваем профиль
          fetchUserProfile(userData)
          return userData
        }
      } catch (e) {
        console.error('Ошибка при загрузке данных пользователя:', e)
      }
    }
    
    return {
      id: null,
      name: '',
      surname: '',
      email: '',
      role: '',
      avatar: null
    }
  }
  
  // Функция для получения актуального профиля пользователя
  const fetchUserProfile = async (userData) => {
    if (!userData || !process.client) return
    
    try {
      const apiBase = useRuntimeConfig().public.apiBase
      console.log('Запрашиваем актуальный профиль...')
      
      // Получаем токен из cookie для ручной передачи
      let token = ''
      if (process.client) {
        const cookies = document.cookie.split(';')
        for (const cookie of cookies) {
          const [name, value] = cookie.trim().split('=')
          if (name === 'token') {
            token = value
            console.log('Найден токен в cookie:', token)
            break
          }
        }
      }
      
      const response = await fetch(`${apiBase}/profile`, {
        method: "GET",
        credentials: "include", // Важно: включаем отправку куки
        headers: {
          "Accept": "application/json",
          "Content-Type": "application/json",
        }
      })
      
      if (response.ok) {
        const profileData = await response.json()
        console.log('Получен актуальный профиль:', profileData)
        
        // Проверяем, что получены все необходимые данные
        if (profileData && profileData.id && profileData.name) {
          // Создаем полностью обновленные данные пользователя
          const updatedData = {
            id: profileData.id,
            name: profileData.name || 'Пользователь',
            surname: profileData.surname || '',
            email: profileData.email || userData.email || '',
            role: profileData.role || userData.role || '',
            avatar: profileData.avatar || userData.avatar || null
          }
          
          console.log('Обновляем данные пользователя:', updatedData)
          
          // Обновляем состояние
          user.value = updatedData
          
          // Сохраняем в localStorage
          if (process.client) {
            localStorage.setItem('user', JSON.stringify(updatedData))
            console.log('Данные пользователя сохранены в localStorage')
          }
          
          return updatedData
        } else {
          console.warn('Получен неполный профиль:', profileData)
        }
      } else {
        console.error('Ошибка при запросе профиля:', response.status, response.statusText)
      }
    } catch (error) {
      console.error('Ошибка при получении профиля пользователя:', error)
    }
    
    // Если произошла ошибка, возвращаем исходные данные
    return userData
  }
  
  // Используем состояние пользователя из Nuxt с более полной информацией
  const user = useState('user', loadUserFromStorage)
  const isLoggedIn = computed(() => !!user.value?.email)
  
  // Проверка роли пользователя
  const hasRole = (role) => {
    return isLoggedIn.value && user.value?.role === role
  }
  
  // Проверка, является ли пользователь организатором
  const isOrganizer = computed(() => hasRole('organizer'))
  
  // Проверка, является ли пользователь участником
  const isParticipant = computed(() => hasRole('participant'))
  
  // Проверка, является ли пользователь модератором
  const isModerator = computed(() => hasRole('moderator'))
  
  // Функция для входа в систему
  const login = async (userData) => {
    console.log('Функция login получила данные:', userData)
    
    // Всегда выполняем запрос к профилю для получения актуальных данных
    let userToSave = {...userData}
    
    try {
      // Делаем небольшую задержку, чтобы убедиться, что cookie установлены
      await new Promise(resolve => setTimeout(resolve, 300))
      
      const apiBase = useRuntimeConfig().public.apiBase
      
      // Получаем токен из cookie для отладки
      if (process.client) {
        const cookies = document.cookie.split(';')
        for (const cookie of cookies) {
          const [name, value] = cookie.trim().split('=')
          if (name === 'token') {
            console.log('Токен в cookie перед запросом профиля:', value)
            break
          }
        }
      }
      
      const response = await fetch(`${apiBase}/profile`, {
        method: "GET",
        credentials: "include", // Явно указываем передачу куки
        headers: {
          "Accept": "application/json",
          "Content-Type": "application/json"
        }
      })
      
      if (response.ok) {
        const profileData = await response.json()
        console.log('Получен профиль в функции login:', profileData)
        
        // Если профиль успешно получен, используем его данные
        if (profileData && profileData.name) {
          userToSave = {
            ...userToSave,
            id: profileData.id || userToSave.id,
            name: profileData.name || userToSave.name,
            surname: profileData.surname || userToSave.surname,
            email: profileData.email || userToSave.email,
            role: profileData.role || userToSave.role,
          }
          console.log('Данные пользователя обновлены из профиля:', userToSave)
        }
      } else {
        console.warn('Не удалось получить профиль при входе, статус:', response.status, response.statusText)
      }
    } catch (error) {
      console.error('Ошибка при получении профиля при входе:', error)
    }
    
    // Сохраняем данные в состоянии
    user.value = userToSave
    
    // Сохраняем пользователя в localStorage
    if (process.client) {
      try {
        localStorage.setItem('user', JSON.stringify(userToSave))
        console.log('Сохранены данные пользователя в localStorage:', userToSave)
      } catch (e) {
        console.error('Ошибка при сохранении данных пользователя:', e)
      }
    }
    
    // Перенаправление на соответствующую страницу
    // Используем setTimeout чтобы дать время на сохранение данных
    setTimeout(() => {
      const redirectPath = useState('redirectPath').value
      if (redirectPath) {
        navigateTo(redirectPath)
        useState('redirectPath').value = null
      } else {
        // Перенаправляем на соответствующую страницу в зависимости от роли
        navigateTo(`/dashboard/${userToSave.role}`)
      }
    }, 100)
  }
  
  // Функция для выхода из системы
  const logout = async () => {
    // Обнуляем состояние пользователя
    user.value = {
      id: null,
      name: '',
      surname: '',
      email: '',
      role: '',
      avatar: null
    }
    
    // Удаляем данные из localStorage
    if (process.client) {
      try {
        localStorage.removeItem('user')
      } catch (e) {
        console.error('Ошибка при удалении данных пользователя:', e)
      }
    }
    
    // Отправляем запрос на выход на сервер
    try {
      const apiBase = useRuntimeConfig().public.apiBase
      await fetch(`${apiBase}/logout`, {
        method: 'POST',
        credentials: 'include'
      })
    } catch (error) {
      console.error('Ошибка при выходе из системы:', error)
    }
    
    // Перенаправляем на страницу входа
    navigateTo('/login')
  }
  
  return {
    user,
    isLoggedIn,
    hasRole,
    isOrganizer,
    isParticipant,
    isModerator,
    login,
    logout
  }
} 