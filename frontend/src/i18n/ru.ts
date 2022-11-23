import { TranslationMessages } from 'react-admin';
import russianMessages from './defaultMessages/russianMessages';

const customRussianMessages: TranslationMessages = {
    ...russianMessages,
    portal: {
        configuration: "Настройки",
        language: "Язык",
        theme: {
            type: "Цветовая схема",
            light: "Светлая",
            dark: "Темная"
        }
    },
    resources: {
        commands: {
            users_name: "Пользователи",
            groups_name: "Группы",
            access_name: "Доступы",
            repository_name: "Репозитории"
        },
        users: {
            name: "Пользователи",
            add_title: "Добавление пользователя",
            edit_title: "Редактирование пользователя",
            fields: {
                login: "Логин",
                name: "Имя пользователя",
                password: "Пароль",
                group: "Группа",
                role: "Роль",
                blocked: "Заблокирован",
                description: "Комментарий"

            }
        },
        groups: {
            name: "Группы",
            edit_title: "Редактирование группы",
            fields: {
                name: "Наименование",
                description: "Комментарий"

            }
        },
        accesses:{
            name: "Управление доступом",
            add_title: "Добавить доступ",
            edit_title: "Изменить",
            fields: {
                name: "Наименование",
                owner_id: "Пользователь",
                resource_type: "Тип ресурса",
                resource_name: "Имя репозитория",
                action:"Вид операции",
                disabled:"Отключен"
            },
            messages:{
                access_tooltip:"Показать доступные репозитории"
            }
        },
        repository:{
            fields:{
                name: "Название",
                size: "Размер",
                tag:"Тэг",
                date:"Дата",
                digest:"Подпись",
                details:"Поднробнее"

            },
            title: "Информация о репозитории",
            tag_list_title: "Список меток (tags)",
            pull_counter: "Количество загрузок: ",
            tag_digest: "Подпись: ",
            tag_media_type: "Тип: ",
            image_platform_details: "Платформа",
            image_config_details: "Параметры",
            image_history_details: "История",
            message_empty_page: "Ни одной записи не найдено. Список репозиториев пуст.",
            message_config_data_not_loading: "Не удалось загрузить данные с конфигурацией образа",
            message_sync_about:"Выполнить синхронизацию данных между RA и Docker Registry",
            message_sync_repo: "Синхронизировать репозитории из реестра",
            message_syncing_repo: "Синхронизация репозиториев запущена",
            message_error_syncing_repo: "Попытка синхронизации завершилась ошибкой",
            message_repo_syncing_running: "Синхронизация уже запущена. Дождитесь окончания окончания операции.",
            button_sync: "Синхронизировать"
        }
    },


};

export default customRussianMessages;
