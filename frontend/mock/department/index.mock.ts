import { toAnyString } from '@/utils'
import { faker } from '@faker-js/faker'
import { SUCCESS_CODE } from '@/constants'

const departmentList: any = []

const citys = [
  'Head office',
  'Northern office',
  'Eastern office',
  'Regional office',
  'Southern office',
  'Western office'
]

for (let i = 0; i < 5; i++) {
  departmentList.push({
    departmentName: citys[i],
    id: toAnyString(),
    createTime: faker.date.anytime(),
    status: faker.number.int({ min: 0, max: 1 }),

    remark: faker.lorem.sentence(),
    children: [
      {
        departmentName: 'Engineering',
        createTime: faker.date.anytime(),

        status: faker.number.int({ min: 0, max: 1 }),
        id: toAnyString(),
        remark: faker.lorem.sentence()
      },
      {
        departmentName: 'Product',
        createTime: faker.date.anytime(),

        status: faker.number.int({ min: 0, max: 1 }),
        id: toAnyString(),
        remark: faker.lorem.sentence()
      },
      {
        departmentName: 'Operations',
        createTime: faker.date.anytime(),

        status: faker.number.int({ min: 0, max: 1 }),
        id: toAnyString(),
        remark: faker.lorem.sentence()
      },
      {
        departmentName: 'Marketing',
        createTime: faker.date.anytime(),

        status: faker.number.int({ min: 0, max: 1 }),
        id: toAnyString(),
        remark: faker.lorem.sentence()
      },
      {
        departmentName: 'Sales',
        createTime: faker.date.anytime(),

        status: faker.number.int({ min: 0, max: 1 }),
        id: toAnyString(),
        remark: faker.lorem.sentence()
      },
      {
        departmentName: 'Support',
        createTime: faker.date.anytime(),

        status: faker.number.int({ min: 0, max: 1 }),
        id: toAnyString(),
        remark: faker.lorem.sentence()
      }
    ]
  })
}

export default [
  {
    url: '/mock/department/list',
    method: 'get',
    response: () => {
      return {
        code: SUCCESS_CODE,
        data: {
          list: departmentList
        }
      }
    }
  },
  {
    url: '/mock/department/table/list',
    method: 'get',
    response: () => {
      return {
        code: SUCCESS_CODE,
        data: {
          list: departmentList,
          total: 5
        }
      }
    }
  },
  {
    url: '/mock/department/users',
    method: 'get',
    timeout: 1000,
    response: ({ query }) => {
      const { pageSize } = query

      const mockList: any = []
      for (let i = 0; i < pageSize; i++) {
        mockList.push({
          username: faker.person.firstName(),

          account: faker.person.lastName(),

          email: faker.internet.email(),

          createTime: faker.date.anytime(),

          id: toAnyString()
        })
      }
      return {
        code: SUCCESS_CODE,
        data: {
          total: 100,
          list: mockList
        }
      }
    }
  },

  {
    url: '/mock/department/user/save',
    method: 'post',
    timeout: 1000,
    response: () => {
      return {
        code: SUCCESS_CODE,
        data: 'success'
      }
    }
  },

  {
    url: '/mock/department/user/delete',
    method: 'post',
    response: ({ body }) => {
      const ids = body.ids
      if (!ids) {
        return {
          code: 500,
          message: 'Select data to delete'
        }
      } else {
        return {
          code: SUCCESS_CODE,
          data: 'success'
        }
      }
    }
  },

  {
    url: '/mock/department/save',
    method: 'post',
    timeout: 1000,
    response: () => {
      return {
        code: SUCCESS_CODE,
        data: 'success'
      }
    }
  },

  {
    url: '/mock/department/delete',
    method: 'post',
    response: ({ body }) => {
      const ids = body.ids
      if (!ids) {
        return {
          code: 500,
          message: 'Select data to delete'
        }
      } else {
        return {
          code: SUCCESS_CODE,
          data: 'success'
        }
      }
    }
  }
]
