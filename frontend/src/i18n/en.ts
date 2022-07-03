import { TranslationMessages } from 'react-admin';
import englishMessages from 'ra-language-english';

const customEnglishMessages: TranslationMessages = {
    ...englishMessages,
    portal: {
        configuration: "Settings",
        language: "Language",
        theme: {
            type: "Theme type",
            light: "Light",
            dark: "Dark"
        }
    },
    resources: {

        commands: {
            users_name: "Users",
            groups_name: "Groups"
        },
        users: {
            name: "Users",
            edit_title: "Edit user entry",
            fields: {
                login: "Login",
                name: "Username",
                password: "Password",
                group: "Group",
                role: "Role",
                blocked: "User blocked",
                description: "Description"

            }
        },
        groups: {
            name: "Groups",
            edit_title: "Edit group",
            fields: {
                name: "Username",
                description: "Description"

            }
        }
    }
};

export default customEnglishMessages;
