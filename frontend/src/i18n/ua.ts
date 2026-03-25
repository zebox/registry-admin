import { TranslationMessages } from 'react-admin';
import ukraineMessages from './defaultMessages/ukraineMessages';

const customUkraineMessages: TranslationMessages = {
    ...ukraineMessages,
    portal: {
        configuration: "Налаштування",
        language: "Мова",
        theme: {
            type: "Колірна схема",
            light: "Світла",
            dark: "Темна"
        }
    },
    resources: {
        commands: {
            users_name: "Користувачі",
            groups_name: "Групи",
            access_name: "Доступи",
            repository_name: "Репозиторії"
        },
        users: {
            name: "Користувачі",
            add_title: "Додавання користувача",
            edit_title: "Редагування користувача",
            fields: {
                login: "Логін",
                name: "Ім'я користувача",
                password: "Пароль",
                group: "Група",
                role: "Роль",
                blocked: "Заблокований",
                description: "Коментар"

            }
        },
        groups: {
            name: "Групи",
            edit_title: "Редагування групи",
            fields: {
                name: "Найменування",
                description: "Коментар"

            }
        },
        accesses:{
            name: "Управління доступом",
            add_title: "Додати доступ",
            edit_title: "Редагування доступу",
            fields: {
                name: "Найменування",
                owner_id: "Користувач",
                resource_type: "Тип ресурсу",
                resource_name: "Ім'я репозиторію",
                action:"Вид операції",
                disabled:"Вимкнено"
            },
            labels: {
                label_special_permission:"Особливі дозволи",
                label_for_all_users: "Анонімні",
                label_for_registered_users: "Тільки зареєстровані"
            },
            messages:{
                access_tooltip:"Показати доступні репозиторії"
            }
        },
        repository:{
            fields:{
                name: "Найменування",
                size: "Розмір",
                tag:"Тег",
                date:"Дата",
                digest:"Підпис",
                details:"Докладніше"

            },
            title: "Інформація про репозиторій",
            tag_list_title: "Список міток (tags)",
            pull_counter: "Кількість завантажень: ",
            tag_digest: "Підпис: ",
            tag_media_type: "Тип: ",
            tag: {
                raw: {
                    details: "Подробиці: ",
                    history: "Історія: ",
                    manifest: "Маніфест: ",
                    layers: "Шари: ",
                    config: "Конфігурація: ",
                    tags: "Теги: ",
                    tag: "Тег: ",
                    os: "ОС: ",
                    arch: "Архітектура: ",
                    schemaVersion:"Версія схеми: ",
                    mediaType:"MediaType: ",
                    total_size:"Загальний розмір: ",
                    digest: "Дайджест: ",
                    content_digest:"Дайджест контенту: ",
                    platform: "Платформа: ",
                    error: "Недоступно: "
                },
            },
            image_platform_details: "Платформа",
            image_config_details: "Параметри",
            image_history_details: "Історія",
            message_empty_page: "Жодного запису не знайдено. Список репозиторіїв порожній.",
            message_config_data_not_loading: "Не вдалося завантажити дані з конфігурацією образу",
            message_sync_about:"Виконати синхронізацію даних між RA та Docker Registry",
            message_sync_repo: "Синхронізувати репозиторії з реєстру",
            message_syncing_repo: "Синхронізацію репозиторіїв запущено",
            message_error_syncing_repo: "Спроба синхронізації завершилась помилкою",
            message_copied_to_clipboard: "Скопійовано",
            message_repo_syncing_running: "Синхронізацію вже запущено. Зачекайте завершення операції.",
            button_sync: "Синхронізувати"
        }
    },


};

export default customUkraineMessages;
