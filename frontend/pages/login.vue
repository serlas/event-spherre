<template>
    <div
        class="min-h-screen flex items-center justify-center relative"
        :style="{ backgroundColor: '#1A1A2E' }"
    >
        <!-- 🌌 ЗВЁЗДНОЕ НЕБО -->
        <div class="absolute inset-0 z-0 pointer-events-none overflow-hidden">
            <div
                v-for="i in 50"
                :key="i"
                class="floating-circle"
                :style="randomStyle()"
            />
        </div>

        <!-- 🟢 ОСНОВНОЙ БЛОК -->
        <div
            class="bg-dark-gray backdrop-blur-md p-10 rounded-2xl shadow-xl w-[500px] border border-[#3A3A50] form-container"
        >
            <h2
                class="text-3xl font-semibold text-center text-[#E4E4F0] font-orbitron mb-8 animate-pulse-slow"
            >
                Вход в систему
            </h2>

            <form @submit.prevent="onLogin" class="space-y-6">
                <input
                    v-model="email"
                    type="email"
                    placeholder="Email"
                    class="w-full p-4 rounded-lg bg-[#3A3A50] placeholder-[#E4E4F0]/70 text-[#E4E4F0] focus:outline-none focus:ring-2 focus:ring-[#6C63FF]/50 transition-all input-neon"
                />
                <input
                    v-model="password"
                    type="password"
                    placeholder="Пароль"
                    class="w-full p-4 rounded-lg bg-[#3A3A50] placeholder-[#E4E4F0]/70 text-[#E4E4F0] focus:outline-none focus:ring-2 focus:ring-[#6C63FF]/50 transition-all input-neon"
                />
                <button
                    type="submit"
                    class="w-full p-4 rounded-lg bg-[#6C63FF] text-[#1A1A2E] font-semibold hover:bg-opacity-90 transition-all ease-in-out shadow-xl hover:shadow-2xl"
                >
                    Войти
                </button>
                
                <!-- Разделитель -->
                <div class="flex items-center my-4">
                    <div class="flex-grow h-px bg-[#3A3A50]"></div>
                    <span class="px-3 text-[#E4E4F0]/70 text-sm">или</span>
                    <div class="flex-grow h-px bg-[#3A3A50]"></div>
                </div>
                
                <!-- Кнопка Google авторизации -->
                <button
                    @click.prevent="handleGoogleLogin"
                    class="w-full p-3 flex items-center justify-center bg-white text-gray-800 rounded-lg font-medium hover:bg-gray-100 transition-all"
                >
                    <svg class="w-5 h-5 mr-2" viewBox="0 0 24 24">
                        <path
                            fill="#4285F4"
                            d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                        />
                        <path
                            fill="#34A853"
                            d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                        />
                        <path
                            fill="#FBBC05"
                            d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                        />
                        <path
                            fill="#EA4335"
                            d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                        />
                    </svg>
                    Войти через Google
                </button>
                
                <div class="flex justify-between mt-4">
                    <NuxtLink
                        to="/register"
                        class="text-[#E4E4F0]/70 hover:text-[#E4E4F0] transition-all"
                    >
                        Регистрация
                    </NuxtLink>
                    <NuxtLink
                        to="/recovery"
                        class="text-[#E4E4F0]/70 hover:text-[#E4E4F0] transition-all"
                    >
                        Забыли пароль?
                    </NuxtLink>
                </div>
            </form>
        </div>

        <!-- ✅ Уведомление -->
        <div
            v-if="showSuccess"
            class="fixed top-6 right-6 bg-[#6C63FF] text-[#1A1A2E] px-6 py-4 rounded-xl shadow-lg animate-fade-slide font-semibold z-50"
        >
            ✅ {{ successMessage || 'Успешный вход!' }}
        </div>

        <!-- ❌ Ошибка входа -->
        <div
            v-if="showError"
            class="fixed top-6 right-6 bg-[#FF6584] text-[#1A1A2E] px-6 py-4 rounded-xl shadow-lg animate-fade-slide font-semibold z-50"
        >
            ❌ {{ errorMessage }}
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRuntimeConfig } from "#app";

const { login } = useAuth();
const email = ref("");
const password = ref("");
const showSuccess = ref(false);
const showError = ref(false);
const csrfToken = ref("");
const errorMessage = ref("");
const successMessage = ref("");

onMounted(async () => {
    console.log("Компонент login.vue монтируется, запускаем fetchCSRFToken");
    await fetchCSRFToken();
});

const fetchCSRFToken = async () => {
    try {
        const url = `${useRuntimeConfig().public.apiBase}/csrf-token`;
        console.log("Отправка запроса на CSRF-токен:", url);
        const response = await $fetch(url, {
            method: "GET",
            credentials: "include",
        });
        console.log("Ответ от /csrf-token:", response);
        if (response.csrf_token) {
            csrfToken.value = response.csrf_token;
            console.log("CSRF token установлен:", csrfToken.value);
        } else {
            console.error("Ответ не содержит csrf_token:", response);
            errorMessage.value = "CSRF-токен отсутствует в ответе";
        }
    } catch (error) {
        console.error("Ошибка получения CSRF-токена:", error.message, error);
        errorMessage.value =
            "Не удалось загрузить CSRF-токен: " + error.message;
    }
};

const onLogin = async () => {
    errorMessage.value = "";
    console.log("Попытка входа, текущий CSRF-токен:", csrfToken.value);
    if (!csrfToken.value) {
        errorMessage.value = "CSRF-токен не загружен";
        console.error("Вход прерван:", errorMessage.value);
        showError.value = true;
        setTimeout(() => {
            showError.value = false;
        }, 2500);
        return;
    }

    try {
        const payload = {
            email: email.value,
            password: password.value,
        };
        console.log("Отправка данных на /login:", payload);
        const response = await $fetch(`${useRuntimeConfig().public.apiBase}/login`, {
            method: "POST",
            credentials: "include",
            headers: {
                "X-CSRF-Token": csrfToken.value,
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
            body: payload,
        });
        console.log("Вход успешен, данные с сервера:", response);
        
        // Формируем базовые данные пользователя из ответа сервера
        const userData = {
            id: response.id || null,
            name: response.name || 'Пользователь', 
            surname: response.surname || '',
            email: email.value,
            role: response.role,
            avatar: null
        };
        
        // Проверяем, есть ли cookie token после входа
        setTimeout(() => {
            const cookies = document.cookie.split(';');
            let hasToken = false;
            for (const cookie of cookies) {
                const [name, value] = cookie.trim().split('=');
                if (name === 'token') {
                    console.log('Токен в cookie после входа:', value);
                    hasToken = true;
                    break;
                }
            }
            if (!hasToken) {
                console.warn('Предупреждение: Cookie с токеном не найден после входа');
            }
        }, 100);
        
        // Пытаемся получить дополнительные данные профиля
        try {
            // Увеличиваем задержку перед запросом профиля для гарантированной синхронизации с сервером
            await new Promise(resolve => setTimeout(resolve, 1000));
            
            const profileResponse = await $fetch(`${useRuntimeConfig().public.apiBase}/profile`, {
                method: "GET",
                credentials: "include", // Важно: включаем отправку куки
                headers: {
                    "Accept": "application/json",
                    "Content-Type": "application/json"
                }
            });
            console.log("Получен профиль с дополнительными данными:", profileResponse);
            
            // Объединяем данные из профиля с базовыми данными
            if (profileResponse) {
                userData.id = profileResponse.id || userData.id;
                userData.name = profileResponse.name || userData.name;
                userData.surname = profileResponse.surname || userData.surname;
                userData.avatar = profileResponse.avatar || userData.avatar;
            }
        } catch (profileError) {
            console.error("Ошибка получения профиля:", profileError);
            
            // Для отладки логируем данные, которые будут использованы
            console.log("Будут использованы базовые данные пользователя:", userData);
        }
        
        // Выполняем вход с собранными данными
        await login(userData);
        
        showSuccess.value = true;
        setTimeout(() => {
            showSuccess.value = false;
        }, 2000);
    } catch (error) {
        const errMsg = error.data?.error || "Неизвестная ошибка сервера";
        console.error("Ошибка входа:", error.message, error);
        errorMessage.value = errMsg;
        showError.value = true;
        setTimeout(() => {
            showError.value = false;
        }, 2500);
    }
};

// Обработка входа через Google (имитация для конкурса)
const handleGoogleLogin = () => {
    // Для демо-целей симулируем успешный вход через Google
    const mockUser = {
        name: "Тестовый Пользователь",
        email: "test@example.com",
        role: "participant"
    };
    
    login(mockUser);
    
    successMessage.value = `Вход через Google выполнен как ${mockUser.name}`;
    showSuccess.value = true;
};

const randomStyle = () => {
    const size = Math.floor(Math.random() * 60) + 20;
    const left = Math.random() * 100;
    const delay = Math.random() * 10;
    const duration = 10 + Math.random() * 10;
    return {
        width: `${size}px`,
        height: `${size}px`,
        left: `${left}%`,
        animationDelay: `${delay}s`,
        animationDuration: `${duration}s`,
    };
};
</script>

<style scoped>
/* 🌫️ АНИМАЦИИ */
@keyframes floatUp {
    0% {
        transform: translateY(0) scale(1);
        opacity: 0.2;
    }
    50% {
        transform: translateY(-50vh) scale(1.2);
        opacity: 0.5;
    }
    100% {
        transform: translateY(-100vh) scale(1);
        opacity: 0;
    }
}
.floating-circle {
    position: absolute;
    background-color: rgba(108, 99, 255, 0.06);
    border-radius: 9999px;
    animation: floatUp linear infinite;
    filter: blur(1px);
}

/* 🌫️ АНИМАЦИИ */
@keyframes fadeIn {
    from {
        opacity: 0;
        transform: translateY(-20px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}
@keyframes fadeSlide {
    from {
        opacity: 0;
        transform: translateY(-20px) translateX(20px);
    }
    to {
        opacity: 1;
        transform: translateY(0) translateX(0);
    }
}

.animate-fade-slide {
    animation: fadeSlide 0.5s ease-out;
}

.animate-fadeIn {
    animation: fadeIn 1s ease-out;
}

@keyframes pulseSlow {
    0%,
    100% {
        opacity: 1;
    }
    50% {
        opacity: 0.7;
    }
}
.animate-pulse-slow {
    animation: pulseSlow 3s infinite;
}

/* 🌙 СТИЛИ */
.bg-dark-blue {
    background-color: #1a1a2e;
}

.bg-dark-gray {
    background-color: #2a2a40;
}

.bg-lighter-gray {
    background-color: #3a3a50;
}

.text-light-gray {
    color: #e4e4f0;
}

.text-pink {
    color: #ff6584;
}

.text-purple {
    color: #6c63ff;
}

button {
    background-color: #6c63ff;
    transition: all 0.3s ease;
}

button:hover {
    background-color: #5b54e6;
}

button:focus {
    outline: none;
    box-shadow: 0 0 8px rgba(108, 99, 255, 0.5);
}

button.shadow-xl {
    box-shadow: 0 4px 8px rgba(108, 99, 255, 0.4);
}

button:hover.shadow-2xl {
    box-shadow: 0 8px 16px rgba(108, 99, 255, 0.6);
}

input {
    background-color: #3a3a50;
    color: #e4e4f0;
    border: 1px solid #3a3a50;
    border-radius: 12px;
    padding: 16px;
    transition: all 0.3s ease;
}

input:focus {
    background-color: #4b4b60;
    border-color: #6c63ff;
    box-shadow: 0 0 6px rgba(108, 99, 255, 0.3);
}

input::placeholder {
    color: #e4e4f0;
}

input.input-neon:focus {
    box-shadow: 0 0 12px rgba(108, 99, 255, 1);
    border-color: #6c63ff;
}

a {
    color: #e4e4f0;
    transition: all 0.3s ease;
}

a:hover {
    color: #6c63ff;
}

/* 🌌 ЗВЁЗДЫ */
.floating-circle {
    position: absolute;
    background-color: rgba(108, 99, 255, 0.1);
    border-radius: 50%;
    animation: floatUp linear infinite;
    filter: blur(1px);
}

@keyframes floatUp {
    0% {
        transform: translateY(0) scale(1);
        opacity: 0.3;
    }
    50% {
        transform: translateY(-50vh) scale(1.3);
        opacity: 0.5;
    }
    100% {
        transform: translateY(-100vh) scale(1);
        opacity: 0;
    }
}

/* 🌟 НЕОНОВАЯ ПОЛОСКА ВОКРУГ ФОРМЫ */
.form-container {
    position: relative;
    padding: 20px;
    box-shadow: 0 0 15px rgba(108, 99, 255, 0.6); /* Неоновый эффект */
    border-radius: 12px;
}
</style>
