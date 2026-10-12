const toUpperCase = (str) => str.charAt(0).toUpperCase() + str.slice(1)

module.exports = {
  description: 'Create vue view',
  prompts: [
    {
      type: 'input',
      name: 'path',
      message: 'Please enter a path',
      default: 'views'
    },
    {
      type: 'input',
      name: 'name',
      message: 'Please enter module name'
    }
  ],
  actions: (data) => {
    const { name, path } = data
    const upperFirstName = toUpperCase(name)

    const actions = []
    if (name) {
      actions.push({
        type: 'add',
        path: `./src/${path}/${upperFirstName}.vue`,
        templateFile: './plop/view/view.hbs',
        data: {
          name,
          upperFirstName
        }
      })
    }

    return actions
  }
}
