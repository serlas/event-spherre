<template>
    <div
        class="min-h-screen flex items-center justify-center relative overflow-hidden"
        :style="{ backgroundColor: '#1A1A2E' }"
    >
        <!-- 🌌 ЗВЕЗДНОЕ НЕБО -->
        <div class="absolute inset-0 z-0 pointer-events-none overflow-hidden">
            <div
                v-for="i in 30"
                :key="i"
                class="floating-circle"
                :style="randomStyle()"
            />
        </div>

        <!-- 🟢 ОСНОВНОЙ БЛОК -->
        <div
            class="relative z-10 bg-[#2A2A40]/90 backdrop-blur-xl border border-[#6C63FF] rounded-3xl p-12 w-[520px] shadow-[0_0_30px_#6C63FF33] animate-fade-up"
        >
            <h2
                class="text-4xl font-bold text-center text-[#E4E4F0] font-orbitron mb-10 tracking-wide animate-pulse-slow"
            >
                Регистрация
            </h2>

            <form @submit.prevent="onRegister" class="space-y-6">
                <transition name="fade-slide" mode="out-in">
                    <input
                        v-model="name"
                        type="text"
                        placeholder="Введите имя"
                        class="input"
                        key="name"
                    />
                </transition>

                <transition name="fade-slide" mode="out-in">
                    <input
                        v-model="surname"
                        type="text"
                        placeholder="Введите Фамилию"
                        class="input"
                        key="surname"
                    />
                </transition>

                <transition name="fade-slide" mode="out-in">
                    <input
                        v-model="email"
                        type="email"
                        placeholder="Email"
                        class="input"
                        key="email"
                    />
                </transition>

                <transition name="fade-slide" mode="out-in">
                    <input
                        v-model="password"
                        type="password"
                        placeholder="Пароль"
                        class="input"
                        key="password"
                    />
                </transition>

                <div class="text-[#E4E4F0] space-y-2">
                    <p class="font-semibold text-sm">Выберите роль:</p>
                    <div class="flex flex-wrap gap-3">
                        <label class="radio-option">
                            <input
                                type="radio"
                                name="role"
                                value="participant"
                                v-model="role"
                            />
                            Участник
                        </label>
                        <label class="radio-option">
                            <input
                                type="radio"
                                name="role"
                                value="organizer"
                                v-model="role"
                            />
                            Организатор
                        </label>
                        <label class="radio-option">
                            <input
                                type="radio"
                                name="role"
                                value="moderator"
                                v-model="role"
                            />
                            Модератор
                        </label>
                    </div>
                </div>

                <button type="submit" class="btn-glow">
                    Зарегистрироваться
                </button>
                
                <!-- Разделитель -->
                <div class="flex items-center my-4">
                    <div class="flex-grow h-px bg-[#3A3A50]"></div>
                    <span class="px-3 text-[#E4E4F0]/70 text-sm">или</span>
                    <div class="flex-grow h-px bg-[#3A3A50]"></div>
                </div>
                
                <!-- Кнопка Google регистрации -->
                <button
                    @click.prevent="handleGoogleRegister"
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
                    Регистрация через Google
                </button>

                <NuxtLink
                    to="/login"
                    class="block text-center text-[#6C63FF] hover:underline mt-4 transition-all hover:text-[#E4E4F0]"
                >
                    Уже есть аккаунт? Войти
                </NuxtLink>
            </form>
        </div>
        
        <!-- Уведомление об успехе -->
        <div
            v-if="showSuccess"
            class="fixed top-6 right-6 bg-[#6C63FF] text-[#1A1A2E] px-6 py-4 rounded-xl shadow-lg animate-fade-slide font-semibold z-50"
        >
            ✅ {{ successMessage || 'Регистрация успешна!' }}
        </div>
        
        <!-- Уведомление об ошибке -->
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
const name = ref("");
const surname = ref("");
const email = ref("");
const password = ref("");
const role = ref("participant");
const csrfToken = ref("");
const errorMessage = ref("");
const showSuccess = ref(false);
const showError = ref(false);
const successMessage = ref("");

onMounted(async () => {
    console.log("Компонент register.vue монтируется");
    await fetchCSRFToken();
});

const fetchCSRFToken = async (attempts = 3, delay = 1000) => {
    for (let i = 0; i < attempts; i++) {
        try {
            const url = `${useRuntimeConfig().public.apiBase}/csrf-token`;
            console.log(`Запрос CSRF-токена (попытка ${i + 1}):`, url);
            const response = await $fetch(url, {
                method: "GET",
                credentials: "include",
            });
            console.log("Ответ от /csrf-token:", response);
            if (response.csrf_token) {
                csrfToken.value = response.csrf_token;
                console.log("CSRF-токен установлен:", csrfToken.value);
                errorMessage.value = "";
                return true;
            } else {
                console.error("Ответ не содержит csrf_token:", response);
                errorMessage.value = "CSRF-токен отсутствует в ответе";
            }
        } catch (error) {
            console.error(
                `Ошибка получения CSRF-токена (попытка ${i + 1}):`,
                error,
            );
            errorMessage.value = `Не удалось загрузить CSRF-токен: ${error.message}`;
            if (i < attempts - 1) {
                console.log(`Повтор через ${delay}мс...`);
                await new Promise((resolve) => setTimeout(resolve, delay));
            }
        }
    }
    console.error("Не удалось получить CSRF-токен");
    errorMessage.value = "Не удалось загрузить CSRF-токен";
    return false;
};

const onRegister = async () => {
    errorMessage.value = "";
    console.log("Попытка регистрации, текущий CSRF-токен:", csrfToken.value);

    // Проверяем токен и запрашиваем снова, если он отсутствует
    if (!csrfToken.value) {
        const success = await fetchCSRFToken();
        if (!success) {
            errorMessage.value = "CSRF-токен не загружен";
            console.error("Регистрация прервана:", errorMessage.value);
            showError.value = true;
            setTimeout(() => {
                showError.value = false;
            }, 2500);
            return;
        }
    }

    try {
        const payload = {
            name: name.value,
            surname: surname.value,
            email: email.value,
            password: password.value,
            role: role.value,
        };
        console.log("Отправка данных на /register:", payload);
        const response = await $fetch(`${useRuntimeConfig().public.apiBase}/register`, {
            method: "POST",
            credentials: "include",
            headers: {
                "X-CSRF-Token": csrfToken.value,
                "Content-Type": "application/json",
            },
            body: payload,
        });
        console.log("Регистрация успешна:", response);
        successMessage.value = "Регистрация успешна!";
        showSuccess.value = true;
        
        // Автоматический логин после успешной регистрации с полными данными пользователя
        await login({
            id: response.id,
            name: name.value,
            surname: surname.value,
            email: email.value,
            role: role.value,
            avatar: null
        });
        
        setTimeout(() => {
            showSuccess.value = false;
        }, 2500);
    } catch (error) {
        const errMsg = error.data?.error || "Неизвестная ошибка сервера";
        console.error("Ошибка регистрации:", error);
        errorMessage.value = errMsg;
        showError.value = true;
        setTimeout(() => {
            showError.value = false;
        }, 2500);
    }
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

// Обработка регистрации через Google (имитация для конкурса)
const handleGoogleRegister = () => {
    // Для демо-целей симулируем успешную регистрацию через Google
    const mockUser = {
        name: "Тестовый Пользователь",
        email: "test@example.com",
        role: "participant"
    };
    
    login(mockUser);
    
    successMessage.value = `Регистрация через Google выполнена как ${mockUser.name}`;
    showSuccess.value = true;
    
    setTimeout(() => {
        showSuccess.value = false;
    }, 2500);
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

@keyframes fadeUp {
    from {
        opacity: 0;
        transform: translateY(40px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}
.animate-fade-up {
    animation: fadeUp 0.8s ease-out forwards;
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

.floating-circle {
    position: absolute;
    background-color: rgba(108, 99, 255, 0.06);
    border-radius: 9999px;
    animation: floatUp linear infinite;
    filter: blur(1px);
}

/* ✨ ПОЛЯ */
.input {
    width: 100%;
    padding: 14px 16px;
    background-color: #1a1a2e;
    border: 1px solid #2a2a40;
    border-radius: 12px;
    color: #e4e4f0;
    font-size: 16px;
    outline: none;
    transition: all 0.3s;
}
.input::placeholder {
    color: #e4e4f088;
}
.input:focus {
    border-color: #6c63ff;
    box-shadow: 0 0 8px #6c63ff66;
}

/* 🎚️ РАДИО */
.radio-option {
    display: flex;
    align-items: center;
    gap: 8px;
    background-color: #1a1a2e;
    padding: 6px 12px;
    border-radius: 10px;
    border: 1px solid #2a2a40;
    cursor: pointer;
    transition: all 0.3s;
}
.radio-option:hover {
    border-color: #6c63ff;
    background-color: #1a1a2e60;
}
input[type="radio"] {
    accent-color: #6c63ff;
}

/* 🌟 КНОПКА */
.btn-glow {
    width: 100%;
    padding: 14px;
    background-color: #6c63ff;
    color: #1a1a2e;
    font-weight: bold;
    font-size: 16px;
    border-radius: 12px;
    transition: all 0.3s;
    box-shadow: 0 0 10px #6c63ff55;
}
.btn-glow:hover {
    background-color: #8b79ff;
    box-shadow: 0 0 20px #6c63ff99;
}

/* КНОПКА "ВЫХОД" */
.btn-exit {
    background-color: #ff6584;
    box-shadow: 0 0 10px #ff658455;
    color: #1a1a2e;
}
.btn-exit:hover {
    background-color: #ff5271;
    box-shadow: 0 0 20px #ff658499;
}

/* ПЛАВНЫЙ ПЕРЕХОД */
.fade-slide-enter-active,
.fade-slide-leave-active {
    transition: all 0.4s ease;
}
.fade-slide-enter-from {
    opacity: 0;
    transform: translateY(20px);
}
.fade-slide-leave-to {
    opacity: 0;
    transform: translateY(-20px);
}
</style>
