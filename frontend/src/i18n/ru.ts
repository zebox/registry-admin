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
                disabled:"Отклюен"
            }
        }
    },


};

export default customRussianMessages;
