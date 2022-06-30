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
            name: "Пользователи"
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
        }
    },
    

};

export default customRussianMessages;
